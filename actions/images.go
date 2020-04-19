package actions

import (
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/templates"
	"net/http"
)

// Страница списка
type ImagesListPage struct {
	Images []models.Image
}

// Страница создания
type ImageCreatePage struct {
	Image   models.Image
	Message string
}

// Регистрация роутов по докер образам
func ImagesRoutes() {
	http.Handle("/docker/images/list", auth.RequireAuthentication(imagesList))
	http.Handle("/docker/images/create", auth.RequireAuthentication(imageCreate))
}

// Страница образов
func imagesList(w http.ResponseWriter, req *http.Request, user models.User) {
	templates.Render(w, "templates/images/list.html", ImagesListPage{
		Images: models.GetImages(),
	}, user)
}

// Страница создания/редактирования образа
func imageCreate(w http.ResponseWriter, req *http.Request, user models.User) {
	imageId := req.URL.Query().Get("imageId")
	image := models.GetImageById(imageId)
	message := ""

	if http.MethodPost == req.Method {
		name := req.FormValue("name")
		description := req.FormValue("description")

		image.Name = name
		image.Description = description
		image.UserId = user.Id

		if image.Save() {
			http.Redirect(w, req, "/docker/images/list", http.StatusSeeOther)
		} else {
			message = "Can't save image. Please try again"
		}
	}

	templates.Render(w, "templates/images/create.html", ImageCreatePage{
		Image:   image,
		Message: message,
	}, user)
}
