package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

func main() {
	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	text := string(bytes)

	slices, ids, spaces := processInput(text)

	messages := getMessages(ids)

	for i := range ids {
		fmt.Printf("%s%s\"body\": %q,", slices[i], spaces[i], messages[i])
	}

	fmt.Printf(slices[len(slices)-1])
}

func processInput(input string) ([]string, []int, []string) {
	slices := getTextSlices(input)
	ids := getIds(input)
	spaces := getSpacesBeforeIds(input)

	return slices, ids, spaces
}

func getTextSlices(input string) []string {
	idsRe := regexp.MustCompile(`[0-9]+,`)
	idsReIndices := idsRe.FindAllStringIndex(input, -1)

	var slices []string

	slices = append(slices, input[0:idsReIndices[0][1]])

	for i := range idsReIndices[1:] {
		slices = append(slices, input[idsReIndices[i][1]:idsReIndices[i+1][1]])
	}

	slices = append(slices, input[idsReIndices[len(idsReIndices)-1][1]:])

	return slices
}

func getIds(input string) []int {
	idsRe := regexp.MustCompile(`[0-9]+`)
	idsReIndices := idsRe.FindAllStringIndex(input, -1)

	var ids []int

	for _, v := range idsReIndices {
		nb, err := strconv.Atoi(input[v[0]:v[1]])
		if err != nil {
			panic(err)
		}

		ids = append(ids, nb)
	}

	return ids
}

func getSpacesBeforeIds(input string) []string {
	spacesRe := regexp.MustCompile(`\s+"id"`)
	spacesReIndices := spacesRe.FindAllStringIndex(input, -1)

	var spaces []string

	for _, v := range spacesReIndices {
		spaces = append(spaces, input[v[0]:v[1]-4])
	}

	return spaces
}

type Data struct {
	ArrayId int
	Body    string `json:"body"`
}

func getMessages(ids []int) []string {
	ch := make(chan Data, len(ids))

	for i, v := range ids {
		go func() {
			resp, err := http.Get(fmt.Sprintf("https://winry.khashaev.ru/posts/%d", v))
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			var rs Data
			json.Unmarshal(body, &rs)
			rs.ArrayId = i

			ch <- rs
		}()
	}

	result := make([]string, len(ids))

	for range result {
		data := <-ch
		result[data.ArrayId] = data.Body
	}

	return result
}
