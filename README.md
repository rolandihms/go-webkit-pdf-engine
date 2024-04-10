## Go API for generating PDF files using the Webkit Headless Engine

A simple Go/Fiber api that handles generating PDF documents form HTML input or URL endpoints.

### Run Using Air

    `air --build.exclude_dir "public"`

... Documentation to be generated

#### URL to PDF

     curl -X POST -H "Content-Type: application/json" --data "{\"title\":\"Test\",\"url\":\"https://nampost.com.na\"}" localhost:3000/url-to-pdf/

    `{
        "status": "success",
        "message": "PDF CReated Successfully",
        "error": "",
        "filename": "test_101031-04-04.pdf"
    }`


#### HTML to PDF

     curl -X POST -H "Content-Type: application/json" --data "{\"title\":\"Test\",\"html\":\"<div>Html is here</div>\"}" localhost:3000/html-to-pdf/

    `{
        "status": "success",
        "message": "PDF CReated Successfully",
        "error": "",
        "filename": "test_101045-04-04.pdf"
    }`




... TO DO:

1. Allow more specific format options, page sizes, dpi, aspect ratios
2. Docker image and multi stage build
3. Compare speed of webkit vs headless chromium.
4. Document the path and including the Webkit binary
5. Implement Templ templating, compare vs mustache and plain HTML
6. Allow option to respond with raw file data instead of the URL

### Issues

Webkit running older version of CSS. No border radius and other limitations
