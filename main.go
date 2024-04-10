package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	wkhtmltopdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

type PDFResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Error    string `json:"error"`
	Filename string `json:"filename"`
}

type PDFrequest struct {
	Html  string `json:"html"`
	Url   string `json:"url"`
	Title string `json:"title"`
}

type RequestPayload struct {
	Html  string `json:"html" xml:"html" form:"html"`
	Url   string `json:"url" xml:"url" form:"url"`
	Title string `json:"title" xml:"title" form:"title"`
}

func main() {

	// Load the .env file in the current directory
	// load .env file
	dotenverr := godotenv.Load(".env")

	if dotenverr != nil {
		log.Fatalf("Error loading .env file")
	}

	//Check if sepcific Prot required to start the server
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}

	// Initialize standard Go html template engine
	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Static("/static", "./public")

	app.Get("/", func(c *fiber.Ctx) error {
		// Render index template
		return c.Render("index", fiber.Map{
			"Title": "Hello, World!",
		})
	})

	//View for when only HTML is provided
	app.Get("/html-view", func(c *fiber.Ctx) error {
		// Render index template
		return c.Render("html-view", fiber.Map{
			"Html": "Hello, World!",
		})
	})

	//Create a POST endpoint for Sending in URL endpoint for the PDF
	app.Post("/url-to-pdf", func(c *fiber.Ctx) error {
		body := new(RequestPayload)

		if err := c.BodyParser(body); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"status":      "error",
				"status_code": "400",
				"message":     "Cannot parse input data",
				"errors":      err.Error(),
			})
		}

		//Assert if url is not empty
		if body.Url == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":      "error",
				"status_code": "400",
				"message":     "Url is required",
			})
		}
		//Assert if Title is not empty
		if body.Title == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":      "error",
				"status_code": "400",
				"message":     "Title is required",
			})
		}
		//Lets generate the PDF
		response := GeneratePDFDocument(PDFrequest{
			Url:   body.Url,
			Title: body.Title,
			Html:  body.Html,
		})
		return c.JSON(response)
	})

	// Create a POST endpoint for Sending in Raw HTML endpoint for the PDF
	app.Post("/html-to-pdf", func(c *fiber.Ctx) error {
		body := new(RequestPayload)

		if err := c.BodyParser(body); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"status":      "error",
				"status_code": "400",
				"message":     "Cannot parse input data",
				"errors":      err.Error(),
			})
		}

		// Assert if url is not empty
		if body.Html == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":      "error",
				"status_code": "400",
				"message":     "Html Input is required",
			})
		}
		// Assert if Title is not empty
		if body.Title == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":      "error",
				"status_code": "400",
				"message":     "Title is required",
			})
		}

		//grab the html inpt and create a static file in teh public directory
		filename := strings.Replace(body.Title, " ", "_", -1) + ".html"
		err := os.WriteFile("./public/"+filename, []byte(body.Html), 0644)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":      "error",
				"status_code": "400",
				"message":     "Error creating HTML file",
				"error":       err.Error(),
			})
		}

		//Set the Url to the created file that is served statically
		body.Url = "http://localhost:" + port + "/static/" + filename

		//Lets generate the PDF
		response := GeneratePDFDocument(PDFrequest{
			Url:   body.Url,
			Title: body.Title,
			Html:  body.Html,
		})
		return c.JSON(response)
	})

	log.Fatal(app.Listen(":" + port))
}

/*
Function to Generate the PDF WIth the WKHTMLTOPDF Webkit engine
*/
func GeneratePDFDocument(req PDFrequest) (res PDFResponse) {

	fmt.Println("WKHTMLTOPDF_PATH:", os.Getenv("WKHTMLTOPDF_PATH"))
	// Create new PDF generator
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		log.Println("Error")
		res := PDFResponse{
			Status:  "error",
			Message: "Error creating PDF",
			Error:   err.Error(),
		}
		return res
	}

	// Set global options
	pdfg.Dpi.Set(72)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.Grayscale.Set(false)

	// Create a new input page from an URL
	page := wkhtmltopdf.NewPage(req.Url)

	// Set options for this page
	page.FooterRight.Set("[page]")
	page.FooterFontSize.Set(10)
	page.Zoom.Set(0.95)

	// Add to document
	pdfg.AddPage(page)

	// Create PDF document in internal buffer
	err = pdfg.Create()
	if err != nil {
		res := PDFResponse{
			Status:  "error",
			Message: "Error creating PDF",
			Error:   err.Error(),
		}
		return res
	}

	//Obtain Timestamp for the current filename
	timestamp := time.Now().Format("2024-01-01")
	//CReate filename and escape special characters from the title
	filename := strings.Replace(req.Title, " ", "_", -1) + "_" + timestamp + ".pdf"
	// Write buffer contents to file on disk
	err = pdfg.WriteFile("./public/" + filename)
	if err != nil {
		res := PDFResponse{
			Status:  "error",
			Message: "Error creating PDF",
			Error:   err.Error(),
		}
		return res
	}
	//Return the response
	return PDFResponse{
		Status:   "success",
		Message:  "PDF CReated Successfully",
		Filename: "public/" + filename,
		Error:    "",
	}

}
