package jobs

import (
	"Forum_BE/services"
	"github.com/robfig/cron/v3"
	"log"
)

func StartCronJobs(qs services.QuestionService) {
	c := cron.New()
	c.AddFunc("0 3 * * *", func() {

		questions, err := qs.SyncQuestionsToRAG()
		if err != nil {
			log.Println("Failed to sync questions to RAG:", err)
			return
		}
		log.Println("Synced", len(questions), "questions to RAG")
	})
	c.Start()
}
