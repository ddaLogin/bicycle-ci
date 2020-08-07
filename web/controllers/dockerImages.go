package controllers

import (
	"github.com/ddalogin/bicycle-ci/auth"
	"github.com/ddalogin/bicycle-ci/models"
	"github.com/ddalogin/bicycle-ci/web/templates"
	"net/http"
)

// Контроллер Docker образов
type DockerImagesController struct {
	auth *auth.Service
}

// Конструктор контроллера Docker образов
func NewImagesController(auth *auth.Service) *DockerImagesController {
	return &DockerImagesController{auth: auth}
}

// Страница списка
type ImagesListPage struct {
	Images []models.DockerImage
}

// Страница создания
type ImageCreatePage struct {
	Image   models.DockerImage
	Message string
}

// Страница образов
func (c *DockerImagesController) List(w http.ResponseWriter, req *http.Request, user models.User) {
	templates.Render(w, "web/templates/images/list.html", ImagesListPage{
		Images: models.GetImages(),
	}, user)
}

// Страница создания/редактирования образа
func (c *DockerImagesController) Create(w http.ResponseWriter, req *http.Request, user models.User) {
	imageId := req.URL.Query().Get("imageId")
	image := models.GetImageById(imageId)
	message := ""

	if http.MethodPost == req.Method {
		image.Name = req.FormValue("name")
		image.Description = req.FormValue("description")
		image.UserId = user.Id

		if image.Save() {
			http.Redirect(w, req, "/docker/images/list", http.StatusSeeOther)
		} else {
			message = "Не удалось создать контейнер, пожалуйста попробуй позже."
		}
	}

	templates.Render(w, "web/templates/images/create.html", ImageCreatePage{
		Image:   image,
		Message: message,
	}, user)
}
