package main

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mohamedkhalaf47/rssAggregator/internal/database"
)

func scraper(
	db *database.Queries,
	concurrency int,
	timeBetweenRequest time.Duration) {
	log.Printf("Scraping on %v goroutines every %s duration", concurrency, timeBetweenRequest)
	ticker := time.NewTicker(timeBetweenRequest)
	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(
			context.Background(),
			int32(concurrency),
		)
		if err != nil {
			log.Printf("Error fetching feeds: %v", err)
			continue
		}
		waitGroup := &sync.WaitGroup{}
		for _, feed := range feeds {
			waitGroup.Add(1)

			go scrapeFeed(db, waitGroup, feed)
		}
		waitGroup.Wait()
	}
}

func scrapeFeed(db *database.Queries, waitGroup *sync.WaitGroup, feed database.Feed) {
	defer waitGroup.Done()

	_, err := db.MarkFeedAsFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("Error marking feed as fetched: %v", err)
		return
	}

	rssFeed, err := urlToFeed(feed.Url)
	if err != nil {
		log.Printf("Error fetching feed from URL %s: %v", feed.Url, err)
		return
	}
	for _, item := range rssFeed.Channel.Item {

		date, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			log.Printf("Error parsing publication date %s for item %s: %v", item.PubDate, item.Title, err)
			continue
		}
		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Description: item.Description,
			PublishedAt: date,
			Url:         item.Link,
			FeedID:      feed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				log.Printf("Post with URL %s already exists, skipping", item.Link)
				continue
			}
			log.Printf("Error creating post for feed %s: %v", feed.Name, err)
			continue
		}
	}
	log.Printf("Feeds %s Collected, %v Posts Found", feed.Name, len(rssFeed.Channel.Item))
}
