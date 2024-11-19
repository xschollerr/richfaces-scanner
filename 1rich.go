package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// xscholler - since 2000-2024
func getPatterns() []string {
	return []string{
		`richfaces`,
		`\.seam`,
		`javax\.faces`,
		`jsf`,
		`faces/javax\.faces`,
		`\.xhtml`,
		`org\.richfaces`,
		`org\.ajax4jsf`,
		`3_3_3\.Finalorg/richfaces/`,
		`3_3_3\.Finalorg/richfaces/renderkit/html/css/basic_classes\.xcss`,
		`3_3_3\.Finalorg/richfaces/renderkit/html/css/extended_classes\.xcss`,
		`3_3_3\.Finalorg\.ajax4jsf\.javascript\.AjaxScript`,
		`RichFaces`,
		`seam/resource`,
		`javax.faces.resource`,
		`JBSEAM`,
		`Seam Application`,
	}
}

func getPaths() []string {
	return []string{
		"/",
		"/index.seam",
		"/index.jsf",
		"/index.faces",
		"/index.xhtml",
		"/login.seam",
		"/login.jsf",
		"/login.faces",
		"/login.xhtml",
		"/home.seam",
		"/home.jsf",
		"/home.faces",
		"/home.xhtml",
		"/main.seam",
		"/main.jsf",
		"/main.faces",
		"/main.xhtml",
		"/app/index.seam",
		"/app/login.seam",
		"/apps/login.seam",
		"/seam/resource/remoting/resource",
		"/faces/javax.faces.resource",
		"/richfaces/renderkit/html/css/basic_classes.xcss",
		"/a4j/g/3_3_3.Finalorg/richfaces/renderkit/html/css/basic_classes.xcss",
	}
}

func getPorts() []string {
	return []string{
		"", // porta padrão (80/443)
		":8080",
		":8443",
		":8181",
		":8000",
		":9090",
	}
}

func compileRegexPatterns(patterns []string) []*regexp.Regexp {
	regexPatterns := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		regexPatterns[i] = regexp.MustCompile(pattern)
	}
	return regexPatterns
}

func createHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("stopped after 5 redirects")
			}
			return nil
		},
	}
}

func normalizeURL(url string) string {
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "http") {
		return "http://" + url
	}
	return url
}

func checkPatterns(content []byte, regexPatterns []*regexp.Regexp) (bool, string) {
	contentStr := string(content)
	for _, regex := range regexPatterns {
		if regex.MatchString(contentStr) {
			return true, regex.String()
		}
	}
	return false, ""
}

func isValidResponse(resp *http.Response) bool {
	if resp.StatusCode == http.StatusNotFound {
		return false
	}

	if resp.StatusCode != http.StatusOK && (resp.StatusCode < 300 || resp.StatusCode > 399) {
		return false
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "text/html") &&
		!strings.Contains(strings.ToLower(contentType), "application/xhtml") &&
		!strings.Contains(strings.ToLower(contentType), "application/xml") {
		return false
	}

	return true
}

func processURL(client *http.Client, baseURL string, regexPatterns []*regexp.Regexp, results chan<- string, processed *sync.Map) {
	ports := getPorts()
	paths := getPaths()

	for _, port := range ports {
		for _, path := range paths {
			url := baseURL + port + path

			if _, exists := processed.LoadOrStore(url, true); exists {
				continue
			}

			resp, err := client.Get(url)
			if err != nil {
				continue
			}

			if !isValidResponse(resp) {
				resp.Body.Close()
				continue
			}

			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()

			if err != nil {
				continue
			}

			found, pattern := checkPatterns(body, regexPatterns)
			if found {
				result := fmt.Sprintf("%s [Status: %d] [Padrão: %s]", url, resp.StatusCode, pattern)
				fmt.Fprintf(os.Stderr, "Encontrado: %s\n", result)
				results <- result
			}
		}
	}
}

func worker(id int, jobs <-chan string, results chan<- string, client *http.Client, regexPatterns []*regexp.Regexp, wg *sync.WaitGroup, processed *sync.Map) {
	defer wg.Done()
	for url := range jobs {
		processURL(client, url, regexPatterns, results, processed)
	}
}

func readURLs(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := normalizeURL(scanner.Text())
		if url != "" {
			urls = append(urls, url)
		}
	}
	return urls, scanner.Err()
}

func saveResults(filename string, results []string) error {
	return os.WriteFile(filename, []byte(strings.Join(results, "\n")), 0644)
}

func main() {
	inputFile := flag.String("i", "", "Arquivo de entrada com a lista de URLs")
	outputFile := flag.String("o", "resultados.txt", "Arquivo de saída para os resultados")
	workers := flag.Int("w", 20, "Número de workers paralelos")
	flag.Parse()

	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Uso: %s -i lista.txt [-o resultados.txt] [-w 20]\n", os.Args[0])
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Iniciando programa...\n")

	urls, err := readURLs(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler arquivo de URLs: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "URLs carregadas: %d\n", len(urls))

	patterns := getPatterns()
	fmt.Fprintf(os.Stderr, "Padrões carregados: %d\n", len(patterns))

	paths := getPaths()
	fmt.Fprintf(os.Stderr, "Caminhos a testar: %d\n", len(paths))

	ports := getPorts()
	fmt.Fprintf(os.Stderr, "Portas a testar: %d\n", len(ports))

	regexPatterns := compileRegexPatterns(patterns)
	client := createHTTPClient()

	processed := &sync.Map{}

	jobs := make(chan string, *workers)
	results := make(chan string, *workers)

	var wg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker(i, jobs, results, client, regexPatterns, &wg, processed)
	}

	go func() {
		for _, url := range urls {
			jobs <- url
		}
		close(jobs)
	}()

	var foundResults []string
	done := make(chan bool)
	go func() {
		for result := range results {
			foundResults = append(foundResults, result)
			fmt.Println(result)
		}
		done <- true
	}()

	wg.Wait()
	close(results)
	<-done

	if len(foundResults) > 0 {
		if err := saveResults(*outputFile, foundResults); err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao salvar resultados: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "\nResultados salvos em: %s\n", *outputFile)
		}
	}

	fmt.Fprintf(os.Stderr, "\nTotal de URLs verificadas: %d\n", len(urls))
	fmt.Fprintf(os.Stderr, "Total de resultados encontrados: %d\n", len(foundResults))
	fmt.Fprintf(os.Stderr, "Programa finalizado.\n")
}
