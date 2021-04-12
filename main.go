package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	proc *exec.Cmd
	cwd string
)

func main() {
	godotenv.Load(".env")
	r := gin.Default()

	cwd, _ = os.Getwd()

	r.GET(`push`, func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.JSON(403, gin.H {
				"status": "error",
				"error": "no code",
			})
			return
		}
		if code != os.Getenv("CODE") {
			c.JSON(403, gin.H {
				"status": "error",
				"error": "wrong code",
			})
			return
		}
		go updateBot()
		c.JSON(200, gin.H {
			"status": "success",
		})
	})
	_, e := os.Stat("./bot")
	if e != nil {
		downloadBot()
	}
	go startBot()
	r.Run(`:8000`)
}

func startBot() {
	runCmd := os.Getenv("RUN_CMD")
	cmd := strings.Split(runCmd, " ")
	args := make([]string, len(cmd) - 2)
	for i := 1; i < len(cmd); i++ {
		args = append(args, cmd[i])
	}
	c := exec.Command(cmd[0], args...)
	c.Dir = cwd + "/bot/"
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Start()
	if err != nil {
		log.Fatalf("Error starting bot: %s", err.Error())
	}
	proc = c
}

func downloadBot() {
	c := exec.Command("git", "clone", os.Getenv("URL"), "bot")
	c.Dir = cwd
	err := c.Run()
	if err != nil {
		log.Fatalf("Error downloading source: %s", err.Error())
	}

	if _, e := os.Stat("./BaseConfig.json"); e != nil {
		log.Fatalf("BaseConfig not found, exiting!")
		os.Exit(-1)
	}
	Copy("./BaseConfig.json", "./bot/BaseConfig.json")
}

func Copy(source, dest string) {
	input, err := ioutil.ReadFile(source)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile(dest, input, 0644)
	if err != nil {
		fmt.Println("Error creating", dest)
		fmt.Println(err)
		return
	}
}

func updateBot() {
	log.Println("Stopping bot")
	if proc != nil {
		err := (*proc.Process).Kill()
		if err != nil {
			log.Printf("Failed to kill process: %s", err.Error())
		}
		s, err := proc.Process.Wait()
		log.Printf("Kill: %t", s.Exited())
		if err != nil {
			log.Printf("Failed to kill process: %s", err.Error())
		}
	}
	_, e := os.Stat("./bot")
	if e != nil {
		downloadBot()
		return
	}

	c := exec.Command("git", "pull")
	cwd, _ := os.Getwd()
	c.Dir = cwd + "/bot/"
	c.Stdout = os.Stdout
	err := c.Run()
	if err != nil {
		log.Fatalf("Error updating source: %s", err.Error())
	}
	if _, e := os.Stat("./BaseConfig.json"); e != nil {
		log.Fatalf("BaseConfig not found, exiting!")
		os.Exit(-1)
	}
	if _, e := os.Stat("./bot/BaseConfig.json"); e != nil {
		Copy("./BaseConfig.json", "./bot/BaseConfig.json")
	}
	log.Println("Starting bot")
	startBot()
}
