package main

import (
	"fmt"
	"log"
	"net/http"
  "os"
	"os/exec"
  "github.com/google/uuid"
  "github.com/golang-jwt/jwt/v5"
)

type Task struct {
  pipeline *Pipeline
}

func removeTempdir(prefix string, tempdir string) {
  cmd := exec.Command("rm", "-rf", tempdir)
  _, err := cmd.Output()
  if err != nil {
    log.Println(prefix, "Failed to remove tempdir:", err)
    return
  }
  log.Println(prefix, "Removed tempdir:", tempdir)
}

func (t *Task) Run(notificationChannel chan<- string) {
  task_uuid := uuid.New()
  prefix := fmt.Sprintf("[%s]", task_uuid)
  log.Println(prefix, "Running task. Pipeline:", t.pipeline.Name)
  notificationChannel <- fmt.Sprintf("Running pipeline: %s", t.pipeline.Name)
  cmd := exec.Command("mktemp", "-d")
  out, err := cmd.Output()
  if err != nil {
    log.Println(prefix, "Failed to create tempdir:", err)
    notificationChannel <- fmt.Sprintf("Failed to create tempdir: %s", err)
    return
  }
  tempdir := string(out)
  log.Println(prefix, "Tempdir path:", tempdir)
  if t.pipeline.Repository != "" {
    notificationChannel <- fmt.Sprintf("Cloning repository")
    private_key_path, err := GetPrivateKeyPath(t.pipeline.Name)
    log.Println(prefix, "Cloning repository:", t.pipeline.Repository)
    cmd = exec.Command("git", "clone", t.pipeline.Repository, tempdir)
    cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=yes -o BatchMode=yes -o IdentitiesOnly=yes -o UserKnownHostsFile=known_hosts", private_key_path))
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    err = cmd.Run()
    if err != nil {
      log.Println(prefix, "Failed to clone repository:", err)
      notificationChannel <- fmt.Sprintf("Failed to clone repository: %s", err)
      removeTempdir(prefix, tempdir)
      return
    }
    log.Println(prefix, "Repository cloned")
  } else {
    log.Println(prefix, "Creating tempdir")
    cmd = exec.Command("mkdir", "-p", tempdir)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    err = cmd.Run()
    if err != nil {
      log.Println(prefix, "Failed to create tempdir:", err)
      notificationChannel <- fmt.Sprintf("Failed to create tempdir: %s", err)
      removeTempdir(prefix, tempdir)
      return
    }
  }
  notificationChannel <- fmt.Sprintf("Running script")
  log.Println(prefix, "Running script")
  cmd = exec.Command("bash", "-c", t.pipeline.Script)
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  cmd.Dir = tempdir
  err = cmd.Run()
  if err != nil {
    log.Println(prefix, "Failed to run script:", err)
    notificationChannel <- fmt.Sprintf("Failed to run script: %s", err)
    removeTempdir(prefix, tempdir)
    return
  }
  log.Println(prefix, "Pipeline finished!")
  notificationChannel <- fmt.Sprintf("Pipeline finished!")
  removeTempdir(prefix, tempdir)
}


func GetPrivateKeyPath(pipeline string) (string, error) {
	err := os.MkdirAll("keys", 0755)
	if err != nil {
		return "", err
	}
	privateKeyPath := fmt.Sprintf("keys/%s", pipeline)
	_, err = os.Stat(privateKeyPath)
	if os.IsNotExist(err) {
		cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", privateKeyPath, "-N", "")
		err := cmd.Run()
		if err != nil {
			return "", err
		}
		publicKeyPath := privateKeyPath + ".pub"
		publicKeyBytes, err := os.ReadFile(publicKeyPath)
		if err != nil {
			return "", err
		}
    log.Println("Public deploy key for pipeline", pipeline, "is:", string(publicKeyBytes))
	}

	return privateKeyPath, nil
}

func main() {
  config, err := LoadConfig("pipelines.yaml")
  if err != nil {
    log.Fatal("Failed to load config:", err)
  }

  if len(os.Args) == 2 && os.Args[1] == "jwt" {
    token := jwt.New(jwt.SigningMethodHS256)
    tokenString, err := token.SignedString([]byte(config.Server.JwtSecret))
    if err != nil {
      log.Fatal("Failed to generate JWT token:", err)
    }
    fmt.Println(tokenString)
    return
  }

  notificationChannel := make(chan string)
  go func() {
    for message := range notificationChannel {
      if config.Server.NotificationCommand == "" {
        continue
      }
      cmd := exec.Command("bash", "-c", config.Server.NotificationCommand)
      cmd.Env = append(os.Environ(), fmt.Sprintf("MESSAGE=%s", message))
      cmd.Stdout = os.Stdout
      cmd.Stderr = os.Stderr
      err := cmd.Run()
      if err != nil {
        log.Println("Failed to send notification:", err)
      }
    }
  }()

  log.Println("Starting server...")

  channel := make(chan Task, config.Server.Workers * 16)

  for i := 0; i < config.Server.Workers; i++ {
    go func() {
      for task := range channel {
        task.Run(notificationChannel)
      }
    }()
  }

  for _, pipeline := range config.Pipelines {
    log.Println("Registering pipeline:", pipeline.Name)
    _, err := GetPrivateKeyPath(pipeline.Name)
    if err != nil {
      log.Fatal("Failed to get private key path for", pipeline.Name, ":", err)
    }
    http.HandleFunc("/" + pipeline.Name, func(w http.ResponseWriter, r *http.Request) {
      pipeline := pipeline
      if r.Method != http.MethodPost {
        log.Println("Invalid method:", r.Method)
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
      }
      tokenString := r.URL.Query().Get("token")
      if tokenString == "" {
        log.Println("Missing authorization header")
        w.WriteHeader(http.StatusUnauthorized)
        return
      }
      token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(config.Server.JwtSecret), nil
      })
      if err != nil {
        log.Println("Failed to parse JWT token:", err)
        w.WriteHeader(http.StatusUnauthorized)
        return
      }
      if !token.Valid {
        log.Println("Invalid JWT token")
        w.WriteHeader(http.StatusUnauthorized)
        return
      }
      task := Task{pipeline: &pipeline}
      channel <- task
      w.WriteHeader(http.StatusOK)
    })
  }

  err = http.ListenAndServe(fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port), nil)
  if err != nil {
    log.Fatal("Failed to start server:", err)
  }
}
