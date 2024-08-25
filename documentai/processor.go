package documentai

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	documentai "cloud.google.com/go/documentai/apiv1"
	"cloud.google.com/go/documentai/apiv1/documentaipb"
	"google.golang.org/api/option"
)

const (
	BASE_LINE_TEXT = "Fecha de último cambio de estado:"
)

type (
	PreLoad struct {
		Previous string
	}

	FileProcessor struct {
	}

	FileRequest struct {
		Content  []byte
		MimeType string
	}

	FileResponse struct {
		RFC            string              `json:"rfc"`
		IDCIF          string              `json:"id_cif"`
		SocialName     string              `json:"social_name"`
		CapitalName    string              `json:"capital_name"`
		CommercialName string              `json:"comercial_name"`
		StartDate      string              `json:"start_date"`
		Status         string              `json:"status"`
		Address        FileAddressResponse `json:"address"`
	}

	FileAddressResponse struct {
		PostalCode    string `json:"postal_code"`
		Street        string `json:"street"`
		Number        string `json:"number"`
		Location      string `json:"location"`
		FederalEntity string `json:"federal_entity"`
		City          string `json:"city"`
	}
)

func NewFileProcessor() *FileProcessor {
	return &FileProcessor{}
}

func (f *FileProcessor) ProcessDocument(request *FileRequest) (*FileResponse, error) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	processorID := os.Getenv("GOOGLE_CLOUD_PROCESSOR_ID")
	credentialsFile := "./credentials.json"

	ctx := context.Background()
	endpoint := fmt.Sprintf("%s-documentai.googleapis.com:443", location)
	client, err := documentai.NewDocumentProcessorClient(ctx, option.WithEndpoint(endpoint), option.WithCredentialsFile(credentialsFile))
	if err != nil {
		fmt.Println(fmt.Errorf("error creating Document AI client: %w", err))
	}
	defer client.Close()

	req := &documentaipb.ProcessRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/processors/%s", projectID, location, processorID),
		Source: &documentaipb.ProcessRequest_RawDocument{
			RawDocument: &documentaipb.RawDocument{
				Content:  request.Content,
				MimeType: request.MimeType,
			},
		},
	}
	resp, err := client.ProcessDocument(ctx, req)

	if err != nil {
		fmt.Println(fmt.Errorf("processDocument: %w", err))
	}
	document := resp.GetDocument()
	text := document.GetText()
	response := scanLine(text)
	response.IDCIF = findByText(text, "idCIF")
	response.Address = FileAddressResponse{}
	response.Address.City = findByText(text, "Nombre del Municipio o Demarcación Territorial")
	response.Address.Number = findByText(text, "Número Exterior")
	response.Address.PostalCode = findByText(text, "Código Postal")
	response.Address.Street = findByText(text, "Y Calle")
	response.Address.Location = findByText(text, "Nombre de la Localidad")
	response.Address.FederalEntity = findByText(text, "Nombre de la Entidad Federativa")

	return response, nil
}

func scanLine(text string) *FileResponse {
	scanner := bufio.NewScanner(strings.NewReader(text))
	found := false
	count := 0
	response := &FileResponse{}

	for scanner.Scan() {
		if found && count <= 5 {
			switch count {
			case 0:
				response.RFC = scanner.Text()
			case 1:
				response.SocialName = scanner.Text()
			case 2:
				response.CapitalName = scanner.Text()
			case 3:
				response.CommercialName = scanner.Text()
			case 4:
				response.StartDate = scanner.Text()
			case 5:
				response.Status = scanner.Text()
			default:
				break
			}
			count++
		}

		if BASE_LINE_TEXT == scanner.Text() {
			found = true
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil
	}

	return response
}

func findByText(fullText string, keyword string) string {
	index := strings.Index(fullText, keyword)
	if index == -1 {
		fmt.Printf("Campo '%s' não encontrado no documento.\n", keyword)
		return ""
	}

	startIndex := index + len(keyword)
	for startIndex < len(fullText) && (fullText[startIndex] == ':') {
		startIndex++
	}

	endIndex := startIndex
	for endIndex < len(fullText) && fullText[endIndex] != '\n' {
		endIndex++
	}

	return fullText[startIndex:endIndex]
}
