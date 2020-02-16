package actions

import (
	"bicycle-ci/auth"
	"bicycle-ci/models"
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

// Регистрация роутов для сборок
func BuildsRoutes() {
	http.Handle("/builds/run", auth.RequireAuthentication(run))
}

// Запуск сборки
func run(w http.ResponseWriter, req *http.Request, user models.User) {
	//projectId := req.URL.Query().Get("projectId")
	//project   := models.GetProjectById(projectId)

	//url := "https://github.com/ddaLogin/testHook.git"

	//cmd := exec.Command("bash", "-c", "git clone https://github.com/ddaLogin/testHook.git")
	//stdout, err := cmd.StdoutPipe()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//if err := cmd.Start(); err != nil {
	//	log.Fatal(err)
	//}
	//out, _ := ioutil.ReadAll(stdout)
	//fmt.Printf("%+v", string(out))

	cmd := exec.Command("bash", "-c", "git clone https://github.com/ddaLogin/testHook.git")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Start()

	buf := make([]byte, 80)
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			fmt.Println(buf[0:n])
		}
		if err != nil {
			break
		}
	}

	cmd.Wait()
}
