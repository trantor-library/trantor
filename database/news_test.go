package database

import "testing"

func TestNews(t *testing.T) {
	const text = "Some news text"

	db := Init(test_host, test_coll)
	defer db.del()

	err := db.AddNews(text)
	if err != nil {
		t.Errorf("db.News(", text, ") return an error: ", err)
	}

	news, err := db.GetNews(1, 1)
	if err != nil {
		t.Fatalf("db.GetNews() return an error: ", err)
	}
	if len(news) < 1 {
		t.Fatalf("No news found.")
	}
	if news[0].Text != text {
		t.Errorf("News text don't match : '", news[0].Text, "' <=> '", text, "'")
	}
}
