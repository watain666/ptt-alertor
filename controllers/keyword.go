package controllers

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/julienschmidt/httprouter"
	"github.com/watain666/ptt-alertor/models"
	"github.com/watain666/ptt-alertor/models/keyword"
)

func KeywordBoards(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	type keywordCount struct {
		board string
		count int
	}
	keywordCounts := make([]keywordCount, 0)
	boards := models.Board().List()
	for _, name := range boards {
		cnt := len(keyword.Subscribers(name))
		kc := keywordCount{board: name, count: cnt}
		keywordCounts = append(keywordCounts, kc)
	}
	sort.Slice(keywordCounts, func(i, j int) bool {
		return keywordCounts[i].count > keywordCounts[j].count
	})
	for _, kc := range keywordCounts {
		fmt.Fprintf(w, "%s: %d\n", kc.board, kc.count)
	}
}
