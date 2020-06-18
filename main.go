package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
)

var prefix string
var suffix string
var length int
var initializeDB bool

var lookups int = 0

func buildDB() DictNode {
	file, _ := ioutil.ReadFile("words.json")
	var result map[string]interface{}
	json.Unmarshal([]byte(file), &result)

	root := DictNode{}

	currentNode := &root

	for word := range result {
		currentNode = &root

		letters := strings.Split(word, "")
		sort.SliceStable(letters, func(i, j int) bool {
			return letters[i] < letters[j]
		})

		for _, letter := range letters {
			if currentNode.Children == nil {
				currentNode.Children = make(map[string]*DictNode, 26)
			}

			if currentNode.Words == nil {
				currentNode.Words = make([]string, 0)
			}

			_, ok := currentNode.Children[letter]
			if !ok {
				//fmt.Println("setting letter", letter)
				currentNode.Children[letter] = &DictNode{}
			}
			currentNode = currentNode.Children[letter]
		}

		currentNode.Words = append(currentNode.Words, word)
		//fmt.Println("word list: ", currentNode.Words)
	}

	return root
}

// DictNode is a node in the dict tree for finding words
type DictNode struct {
	Words    []string
	Children map[string]*DictNode
}

func (dict *DictNode) printVisit() {
	if dict.Words != nil {
		fmt.Println("Visiting: ", dict.Words)
	} else {
		fmt.Println("Visiting node (wordless)")
	}
	fmt.Print("Children: ")
	for letter := range dict.Children {
		fmt.Print(letter, " ")
	}
	fmt.Println()
}

func uniq(words []string) []string {
	// fmt.Println("words for uniq: ", words, len(words))
	wordmap := make(map[string]int, 0)
	for _, word := range words {
		wordmap[word] = 1
	}

	wordlist := make([]string, 0)
	for word := range wordmap {
		wordlist = append(wordlist, word)
	}

	return wordlist
}

func sortForWWF(words []string) []string {
	sort.SliceStable(words, func(i, j int) bool {
		return len(words[i]) > len(words[j])
	})
	return words
}

func splice(array []string, index int) []string {
	arr := append([]string{}, array[:index]...)
	arr = append(arr, array[index+1:]...)
	return arr
}

func (dict *DictNode) getCandidates(letters []string) []string {
	lookups += 1
	candidates := dict.Words
	//fmt.Println("checking letters: ", dict.Words, letters)

	if len(letters) == 0 {
		//fmt.Println("0 letters, candidates: ", candidates)
		return candidates
	}

	for idx, letter := range letters {
		if letter == "_" {
			for _, child := range dict.Children {
				candidates = append(candidates, child.getCandidates(splice(letters, idx))...)
			}
		}

		node, ok := dict.Children[letter]
		if ok {
			//node.printVisit()
			//fmt.Println("splitting letters: ", letters[:idx], letters[idx+1:], idx)
			candidates = append(candidates, node.getCandidates(splice(letters, idx))...)
		}
	}

	return candidates
}

func (dict *DictNode) saveDB() {
	var db bytes.Buffer

	enc := gob.NewEncoder(&db)
	err := enc.Encode(dict)
	if err != nil {
		log.Fatal("encode:", err)
	}
	ioutil.WriteFile("words.db", db.Bytes(), 0664)
}

func loadDB() DictNode {
	raw, err := ioutil.ReadFile("words.db")
	if err != nil {
		panic("couldn't load db")
	}

	dec := gob.NewDecoder(bytes.NewBuffer(raw))

	var dict DictNode
	err = dec.Decode(&dict)
	if err != nil {
		log.Fatal("decode error:", err)
	}

	return dict
}

func matchesPrefix(word string) bool {
	if prefix != "" {
		return strings.HasPrefix(word, prefix)
	}
	return true
}

func matchesSuffix(word string) bool {
	if suffix != "" {
		return strings.HasSuffix(word, suffix)
	}
	return true
}

func matchesLength(word string) bool {
	if length > 0 {
		return len(word) == length
	}

	return true
}

func applyFlags(words []string) []string {
	filtered := make([]string, 0)
	for _, word := range words {
		if matchesPrefix(word) && matchesSuffix(word) && matchesLength(word) {
			filtered = append(filtered, word)
		}
	}
	return filtered
}

/*
Interactions
> set jbhwokd # sets available letters
> -xx 				# applies suffix -xx
> xx-					# applies prefix xx-
> -xx-				# finds words with xx in the middle
> 3						# lists words that are 3 letters long
> 3 x-				# lists words that are 3 letters long that start with x
> help				# shows this!
*/

func applyFilter(words []string, filter string) []string {
	filtered := make([]string, 0)
	filterText := strings.ReplaceAll(filter, "-", "")

	// fmt.Println("raw words: ", words)

	for _, word := range words {
		matches := word != filterText
		if strings.HasPrefix(filter, "-") && strings.HasSuffix(filter, "-") {
			matches = matches && strings.Contains(word, filterText)
		} else if strings.HasPrefix(filter, "-") {
			matches = matches && strings.HasSuffix(word, filterText)
		} else if strings.HasSuffix(filter, "-") {
			matches = matches && strings.HasPrefix(word, filterText)
		}

		if matches {
			filtered = append(filtered, word)
		}
	}

	return filtered
}

func nav() {
	reader := bufio.NewReader(os.Stdin)
	node := loadDB()
	for {
		node.printVisit()
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		next := strings.Split(text, "")[0]
		node = *node.Children[next]
	}
}

func shell() {
	reader := bufio.NewReader(os.Stdin)
	db := loadDB()
	fmt.Println("WWF Shell for Champs")
	letters := ""

	for {
		lookups = 0
		filter := ""
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		tokenized := strings.Split(text, " ")

		if tokenized[0] == "set" {
			letters = tokenized[1]
		} else {
			filter = text
		}

		// splitLetters := strings.Split(letters+suffix+prefix, "")
		splitLetters := strings.Split(letters+strings.ReplaceAll(filter, "-", ""), "")

		//fmt.Println("splitLetters: ", splitLetters)
		sort.SliceStable(splitLetters, func(i, j int) bool {
			return splitLetters[i] < splitLetters[j]
		})

		candidates := sortForWWF(uniq(db.getCandidates(splitLetters)))
		// candidates = applyFlags(candidates)

		filtered := applyFilter(candidates, filter)
		fmt.Println(len(filtered), " matches")
		fmt.Println("looked at ", lookups, " nodes")
		fmt.Println(filtered)
	}
}

func main() {
	var runShell bool
	var explore bool
	flag.StringVar(&prefix, "prefix", "", "prefix to find")
	flag.StringVar(&suffix, "suffix", "", "suffix to find")
	flag.IntVar(&length, "length", -1, "length of word")
	flag.BoolVar(&initializeDB, "init", false, "whether or not to create db")
	flag.BoolVar(&runShell, "shell", true, "whether to run in shell mode")
	flag.BoolVar(&explore, "explore", false, "whether to navigate the db")

	flag.Parse()

	if initializeDB {
		db := buildDB()
		db.saveDB()
		fmt.Println("DB initialized")
		os.Exit(0)
	}

	if explore {
		nav()
		os.Exit(0)
	}

	if runShell {
		shell()
		os.Exit(0)
	}

	letters := os.Args[len(os.Args)-1]

	splitLetters := strings.Split(letters+suffix+prefix, "")

	db := loadDB()

	candidates := sortForWWF(uniq(db.getCandidates(splitLetters)))
	candidates = applyFlags(candidates)

	fmt.Println("suggestions: ", candidates, len(candidates))
}
