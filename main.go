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

	// 	input := `{
	//   "id": 1,
	//   "replies": [
	//     {
	//       "id": 2,
	//       "replies": []
	//     },
	//     {
	//       "id": 3,
	//       "replies": [
	//         {
	//           "id": 4,
	//           "replies": []
	//         },
	//         {
	//           "id": 5,
	//           "replies": []
	//         }
	//       ]
	//     }
	//   ]
	// }`

	// scanner := bufio.NewScanner(os.Stdin)
	// scanner.Scan()
	// text := scanner.Text()
	// fmt.Println(text)

	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	text := string(bytes)
	// fmt.Println(text)

	slices, ids, spaces := processInput(text)

	messages := getMessages(ids)

	for i := range ids {
		fmt.Printf("%s%s\"body\": %q,", slices[i], spaces[i], messages[i])
	}
	fmt.Printf(slices[len(slices)-1])
}

func processInput(input string) ([]string, []int, []string) {

	re2 := regexp.MustCompile(`[0-9]+,`)
	ind2 := re2.FindAllStringIndex(input, -1)

	var slices []string

	slices = append(slices, input[0:ind2[0][1]])

	for i := range ind2[1:] {
		slices = append(slices, input[ind2[i][1]:ind2[i+1][1]])
	}

	slices = append(slices, input[ind2[len(ind2)-1][1]:])

	var ids []int

	for _, v := range ind2 {
		nb, err := strconv.Atoi(input[v[0] : v[1]-1])
		if err != nil {
			panic(err)
		}

		ids = append(ids, nb)
	}

	re3 := regexp.MustCompile(`\s+"id"`)
	ind3 := re3.FindAllStringIndex(input, -1)

	spaces := make([]string, len(ids))

	for i, v := range ind3 {
		spaces[i] = input[v[0] : v[1]-4]
	}

	return slices, ids, spaces
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

type Data struct {
	ArrayId int
	Body    string `json:"body"`
}
