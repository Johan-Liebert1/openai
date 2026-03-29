package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/asticode/go-astisub"
	"github.com/davecgh/go-spew/spew"
)

type SubsFileType = string

const (
	SubsFileTypeSRT = "srt"
	SubsFileTypeASS = "ass"
)

const SaveAsASS = true

var SpewPrinter = spew.ConfigState{Indent: "    ", MaxDepth: 5}
var inputSubsFileName string

func sendOpenAIRequest(body OpenAIAPIRequest) (GPTResponse, error) {
	gptResponse := GPTResponse{}

	bodyBytes, err := json.Marshal(body)

	if err != nil {
		return gptResponse, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)

	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.openai.com/v1/chat/completions",
		bytes.NewBuffer(bodyBytes),
	)

	if err != nil {
		return gptResponse, fmt.Errorf("Error creating request: %+v\n", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("API_KEY")))

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return gptResponse, fmt.Errorf("Error making request: %+v\n", err)
	}

	fmt.Printf("StatusCode: %d\n", resp.StatusCode)

	// size := int64(math.Pow(2, 23))
	// respBody := make([]byte, size)

	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return gptResponse, fmt.Errorf("Error reading response: %+v\n", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		err = json.Unmarshal(respBody, &gptResponse)

		if err != nil {
			SpewPrinter.Dump(respBody)
			return gptResponse, fmt.Errorf(
				"Error unmarshalling response into GPTResponse: %+v\n",
				err,
			)
		}

		// SpewPrinter.Dump(gptResponse)

		return gptResponse, nil
	}

	response := map[string]any{}

	err = json.Unmarshal(respBody, &response)

	if err != nil {
		SpewPrinter.Dump(response)
		return gptResponse, fmt.Errorf("Error unmarshalling response to map: %+v\n", err)
	}

	SpewPrinter.Dump(response)

	return gptResponse, fmt.Errorf("Welp")
}

func getSubtitles(path string) *astisub.Subtitles {
	subtitles, err := astisub.OpenFile(path)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	return subtitles
}

func getJSONArray(subtitles *astisub.Subtitles, start, end int) string {
	s := "["

	for _, item := range subtitles.Items[start:end] {
		s += fmt.Sprintf(`"%s",`, item.String())
	}

	s += "]"

	return s
}

func writeSubsStringArrayToFile(subs []string, start, batch int) {
	SpewPrinter.Dump(subs)

	f, err := os.OpenFile(
		fmt.Sprintf("subsStrings-start-%d-end-%d.json", start, batch),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0644,
	)

	if err != nil {
		fmt.Printf("Failed to open file. Err: %+v\n", err)
		return
	}

	defer f.Close()

	b, err := json.Marshal(subs)

	if err != nil {
		fmt.Printf("Failed to marshal to json. Err: %+v\n", err)
		return
	}

	_, err = f.Write(b)

	if err != nil {
		fmt.Printf("Failed to write to file. Err: %+v\n", err)
		return
	}
}

func fixJson(content string) string {
	start := strings.Index(content, "[")

	i := len(content) - 1
	end := -1

	for i >= 0 {
		if content[i] == ']' {
			end = i
		}

		i--
	}

	return content[start : end+1]
}

func createNewSubsFile(
	subs *astisub.Subtitles,
	newSubtitlesStringArray []string,
	filetype SubsFileType,
) {
	newSubs := make([]*astisub.Item, len(subs.Items))
	copy(newSubs, subs.Items)

	for i := range len(newSubs) {
		lines := []astisub.Line{}

		for _, line := range strings.Split(newSubtitlesStringArray[i], "\n") {
			lines = append(lines, astisub.Line{
				Items: []astisub.LineItem{
					{Text: line},
				},
			})
		}

		newSubs[i].Lines = lines
	}

	s := astisub.Subtitles{Items: newSubs, Metadata: &astisub.Metadata{}}
	s.Write(fmt.Sprintf("./newthing.%s", filetype))
}

// returns whether to exit program or not
func handleArgs() bool {
	arg := os.Args[1]

	switch arg {
	case "write":
		// Need a subs json file and a subtitles file and writes a new file
		{
			if len(os.Args) != 4 {
				fmt.Printf("Usage: ./exe write <original subs file> <subs json file>\n")
				return true
			}

			subs := getSubtitles(os.Args[2])

			bytes, err := os.ReadFile(os.Args[3])

			if err != nil {
				fmt.Printf("Failed to read file '%s', with err: %+v\n", os.Args[3], err)
				return true
			}

			newSubtitlesStringArray := []string{}

			err = json.Unmarshal(bytes, &newSubtitlesStringArray)

			if err != nil {
				fmt.Printf(
					"Failed to unmarshal contents of file '%s', with err: %+v\n",
					os.Args[3],
					err,
				)
				return true
			}

			fileType := SubsFileTypeSRT

			idx := strings.LastIndex(inputSubsFileName, ".")

			if inputSubsFileName[idx+1:] == SubsFileTypeASS || SaveAsASS {
				fileType = SubsFileTypeASS
			}

			createNewSubsFile(subs, newSubtitlesStringArray, fileType)
		}

	case "convert":
		{
			if len(os.Args) != 3 {
				fmt.Printf("Usage: ./exe convert <original subs file>\n")
				return true
			}

			subs := getSubtitles(os.Args[2])

			fname := strings.LastIndex(os.Args[2], ".")
			outputFname := os.Args[2][:fname] + ".ass"

			subs.Metadata = &astisub.Metadata{}

			subs.Write(fmt.Sprintf("%s", outputFname))
		}

	case "separate-jap-eng":
		{
			if len(os.Args) < 3 {
				fmt.Println("Need a file name")
				return true
			}

			fileName := os.Args[2]

			subs := getSubtitles(fileName)

			japLines := []string{}
			engLines := []string{}

			for idx, item := range subs.Items {
				jap, eng := item.Lines[0], item.Lines[1]
				japLines = append(japLines, fmt.Sprintf("%d\t%s", idx, jap.String()))
				engLines = append(engLines, fmt.Sprintf("%d\t%s", idx, eng.String()))
			}

			outFileName := fmt.Sprintf("%s.output", fileName)
			file, err := os.Create(outFileName)

			if err != nil {
				fmt.Printf("Err: %+v\n", err)
				return true
			}

			_, err = file.WriteString(strings.Join(japLines, "\n"))
			if err != nil {
				fmt.Printf("Err: %+v\n", err)
				return true
			}

			_, err = file.WriteString(strings.Join(engLines, "\n"))
			if err != nil {
				fmt.Printf("Err: %+v\n", err)
				return true
			}

			fmt.Printf("Wrote to %s\n", outFileName)

			// Always exit the program
			return true
		}

	default:
		{
			inputSubsFileName = arg
			return false
		}
	}

	return true
}

func main() {
	if len(os.Args) > 1 {
		if handleArgs() {
			return
		}
	}

	if Romaji {
		DevPrompt = DevPromptRomaji
	} else {
		DevPrompt = DevPromptNonRomaji
	}

	chatMessages = []RequestMessage{
		{
			Role:    "developer",
			Content: DevPrompt,
		},
	}

	isJson := false

	batchSize := 15
	currentBatch := 0

	subs := getSubtitles(inputSubsFileName)

	if subs == nil {
		return
	}

	newSubtitlesStringArray := []string{}

	errCount := 0
	retriesThreshold := 5

	for currentBatch < len(subs.Items) {
		start := currentBatch * batchSize
		end := currentBatch*batchSize + batchSize

		if start >= len(subs.Items) && isJson {
			break
		}

		fmt.Printf("CurrentBatch: %d / %d\n", currentBatch, len(subs.Items))

		text := ""

		if isJson {
			text = getJSONArray(
				subs,
				start,
				min(end, len(subs.Items)),
			)
		} else {
			text = subs.Items[currentBatch].String()
		}

		openAiApiReq := OpenAIAPIRequest{
			Model: "gpt-4o-mini-2024-07-18",
			Store: true,
			Messages: GetConverstaionMessages(RequestMessage{
				Role:    "user",
				Content: text,
			}),
		}

		// SpewPrinter.Dump(openAiApiReq.Messages)

		resp, err := sendOpenAIRequest(openAiApiReq)

		if err != nil {
			errCount++

			SpewPrinter.Dump(resp)
			fmt.Println(err)

			if errCount >= retriesThreshold {
				writeSubsStringArrayToFile(newSubtitlesStringArray, start-batchSize, start-1)
				return
			}

			continue
		}

		errCount = 0

		content := resp.Choices[0].Message.Content

		// Also send the assistant message back as apperantly it can lose context
		chatMessages = append(chatMessages, RequestMessage{
			Role:    "assistant",
			Content: content,
		})

		if isJson {
			content = fixJson(content)
			newSubsString := []string{}
			err = json.Unmarshal([]byte(content), &newSubsString)

			if err != nil {
				writeSubsStringArrayToFile(newSubtitlesStringArray, start-batchSize, start-1)
				fmt.Printf(
					"Failed to unmarshal GPT response. Err: %+v, Batch: %d\n",
					err,
					currentBatch,
				)
				return
			}

			newSubtitlesStringArray = append(newSubtitlesStringArray, newSubsString...)
		} else {
			newSubtitlesStringArray = append(newSubtitlesStringArray, content)
		}

		fmt.Printf("'%s'\n", content)

		currentBatch++

		if !isJson {
			time.Sleep(200 * time.Millisecond)
		}
	}

	writeSubsStringArrayToFile(newSubtitlesStringArray, currentBatch*batchSize, len(subs.Items))

	fileType := SubsFileTypeSRT

	idx := strings.LastIndex(inputSubsFileName, ".")

	if inputSubsFileName[idx+1:] == SubsFileTypeASS || SaveAsASS {
		fileType = SubsFileTypeASS
	}

	createNewSubsFile(subs, newSubtitlesStringArray, fileType)
}
