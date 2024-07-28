package main

import (
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*html")),
	}
}

type Count struct {
	Count int
}

var id = 0

type Contact struct {
	ID    int
	Name  string
	Email string
}

func newContact(name, email string) Contact {
	id++

	return Contact{ID: id, Name: name, Email: email}
}

type Contacts = []Contact

func (d *Data) indexOf(id int) int {
	for i, contact := range d.Contacts {
		if contact.ID == id {
			return i
		}
	}

	return -1
}

func (d *Data) emailExists(email string) bool {
	for _, contact := range d.Contacts {
		if contact.Email == email {
			return true
		}
	}

	return false
}

type Data struct {
	Contacts Contacts
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

func newData() Data {
	return Data{
		Contacts: []Contact{
			newContact("test", "test@email"),
			newContact("test2", "test+2@email"),
			newContact("test3", "test+3@email"),
		},
	}
}

type Page struct {
	Data Data
	Form FormData
}

func newPage() Page {
	return Page{
		Data: newData(),
		Form: newFormData(),
	}
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Renderer = newTemplate()

	e.Static("/images", "images")
	e.Static("/css", "css")

	page := newPage()

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", page)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		if page.Data.emailExists(email) {
			formData := newFormData()
			formData.Values["name"] = name
			formData.Values["email"] = email
			formData.Errors["email"] = "Email already exists"
			return c.Render(http.StatusUnprocessableEntity, "form", formData)
		}

		contact := newContact(name, email)
		page.Data.Contacts = append(page.Data.Contacts, contact)

		c.Render(http.StatusOK, "form", newFormData())
		return c.Render(http.StatusOK, "oob-contact", contact)
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {
		time.Sleep(3 * time.Second)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.String(http.StatusNotFound, "Contact not found")
		}

		index := page.Data.indexOf(id)
		page.Data.Contacts = append(page.Data.Contacts[:index], page.Data.Contacts[index+1:]...)

		return c.NoContent(http.StatusOK)
	})

	e.Logger.Fatal(e.Start(":8000"))
}
