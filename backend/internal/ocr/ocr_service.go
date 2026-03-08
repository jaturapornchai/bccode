package ocr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type IOcrService interface {
	UploadOcr(resourceKey string, urlResources []string) ([]map[string]interface{}, error)
	ResultOcr(resourceKey string, urlResources []string) ([]map[string]interface{}, error)
}

type OcrService struct {
	apiUrl string
	apiKey string
}

func NewOcrService() OcrService {
	insSvc := OcrService{
		apiUrl: "https://manager.ztrus.com/api/document",
		apiKey: "x2PW1PXG::8749bc621266976d992166c3a001f83f",
	}
	return insSvc
}

func (svc OcrService) UploadOcr(resourceKey string, urlResources []string) ([]map[string]interface{}, error) {

	urlApi := fmt.Sprintf("%s/upload", svc.apiUrl)

	result := []map[string]interface{}{}

	wg := sync.WaitGroup{}
	for idx, urlResource := range urlResources {
		wg.Add(1)

		go func(idx int, urlResource string) {
			trackingID := fmt.Sprintf("%s-%d", resourceKey, idx+1)

			fileContent, filename, err := svc.downloadFileFromURL(urlResource)
			if err != nil {
				fmt.Println("Error downloading file:", err)

			}

			resultUpload, _ := svc.postFile(urlApi, OcrUpload{
				TrackingID: trackingID,
				FormIndex:  0,
			}, FileContent{
				FileName: filename,
				Content:  fileContent,
			})

			resultUpload["tracking_id"] = trackingID
			result = append(result, resultUpload)
			wg.Done()
		}(idx, urlResource)
	}
	wg.Wait()

	return result, nil
}

func (svc OcrService) ResultOcr(resourceKey string, urlResources []string) ([]map[string]interface{}, error) {

	urlApi := fmt.Sprintf("%s/result", svc.apiUrl)

	result := []map[string]interface{}{}

	wg := sync.WaitGroup{}
	for idx, urlResource := range urlResources {
		wg.Add(1)
		go func(idx int, urlResource string) {

			trackingID := fmt.Sprintf("%s-%d", resourceKey, idx+1)

			fileContent, filename, err := svc.downloadFileFromURL(urlResource)
			if err != nil {
				fmt.Println("Error downloading file:", err)

			}

			resultUpload, _ := svc.postResult(urlApi, OcrResault{
				TrackingID:    trackingID,
				Type:          "json",
				Url:           1,
				RawHeader:     1,
				Confident:     1,
				SignatureCode: 1,
				Startdate:     "",
				Stopdate:      "",
			}, FileContent{
				FileName: filename,
				Content:  fileContent,
			})

			resultUpload["tracking_id"] = trackingID
			result = append(result, resultUpload)
			wg.Done()
		}(idx, urlResource)
	}
	wg.Wait()

	return result, nil
}

func (svc OcrService) downloadFileFromURL(url string) (io.Reader, string, error) {
	// Get the response bytes from the url
	response, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()

	// Get the filename from the Content-Disposition header or fallback to a default
	contentDisposition := response.Header.Get("Content-Disposition")
	filename := "default_filename"
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			filename = params["filename"]
		}
	}

	if filename == "default_filename" {
		temp1 := strings.Split(url, "?")[0]
		temp2 := strings.Split(temp1, "/")
		filename = temp2[len(temp2)-1]

	}

	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, response.Body)
	return &buffer, filename, err
}

func (svc OcrService) postFile(url string, ocrUpload OcrUpload, fileContent FileContent) (map[string]interface{}, error) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	writer.WriteField("tracking_id", ocrUpload.TrackingID)
	writer.WriteField("form_index", strconv.Itoa(int(ocrUpload.FormIndex)))

	formFile, err := writer.CreateFormFile("file[0]", fileContent.FileName)
	if err != nil {
		return map[string]interface{}{}, err
	}
	_, err = io.Copy(formFile, fileContent.Content)
	if err != nil {
		return map[string]interface{}{}, err
	}

	// Close the writer before sending the request
	err = writer.Close()
	if err != nil {
		return map[string]interface{}{}, err
	}

	// Create a new POST request
	req, err := http.NewRequest(http.MethodPost, url, &buffer)
	if err != nil {
		return map[string]interface{}{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("x-api-key", svc.apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{}, err
	}
	defer resp.Body.Close()

	resultBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{}, err
	}

	resultJson := map[string]interface{}{}

	err = json.Unmarshal(resultBytes, &resultJson)
	if err != nil {
		return map[string]interface{}{}, err
	}

	// if resp.StatusCode != http.StatusOK {
	// 	return map[string]interface{}{}, fmt.Errorf("bad status: %s", resp.Status)
	// }

	return resultJson, nil
}

func (svc OcrService) postResult(url string, ocrResault OcrResault, fileContent FileContent) (map[string]interface{}, error) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	writer.WriteField("tracking_id", ocrResault.TrackingID)
	writer.WriteField("type", ocrResault.Type)
	writer.WriteField("url", strconv.Itoa(int(ocrResault.Url)))
	writer.WriteField("raw_header", strconv.Itoa(int(ocrResault.RawHeader)))
	writer.WriteField("confident", strconv.Itoa(int(ocrResault.Confident)))
	writer.WriteField("signature_code", strconv.Itoa(int(ocrResault.SignatureCode)))
	writer.WriteField("startdate", ocrResault.Startdate)
	writer.WriteField("stopdate", ocrResault.Stopdate)

	// Create a new POST request
	req, err := http.NewRequest(http.MethodPost, url, &buffer)
	if err != nil {
		return map[string]interface{}{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("x-api-key", svc.apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{}, err
	}
	defer resp.Body.Close()

	resultBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{}, err
	}

	resultJson := map[string]interface{}{}

	err = json.Unmarshal(resultBytes, &resultJson)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return resultJson, nil
}
