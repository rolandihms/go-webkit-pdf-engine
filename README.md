### Go API for generating PDF files using the Webkit Headless Engine

A simple Go/Fiber api that handles generating PDF documents form HTML input or URL endpoints.


# Run Using Air

    `air --build.exclude_dir "public"`

... Documentation to be generated

... TO DO:

1. Allow more specific format options, page sizes, dpi, aspect ratios
2. Docker image and multi stage build
3. Compare speed of webkit vs headless chromium.
4. Document the path and including the Webkit binary
5. Implement Templ templating, compare vs mustache and plain HTML
6. Allow option to respond with raw file data instead of the URL

### Issues
Webkit running older version of CSS. No border radius and other limitations