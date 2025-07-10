package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-github/v58/github"
	"golang.org/x/oauth2"
)

func main() {
	// Hardcoded GitHub PAT (⚠️ Replace this with your token)
	token := "" // Replace with your token

	org := "" // Replace with your GitHub organization name

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	var allMembers []*github.User
	opts := &github.ListMembersOptions{ListOptions: github.ListOptions{PerPage: 50}}

	for {
		members, resp, err := client.Organizations.ListMembers(ctx, org, opts)
		if err != nil {
			log.Fatalf("Failed to list members: %v", err)
		}
		allMembers = append(allMembers, members...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	cutoff := time.Now().AddDate(0, 0, -90)
	activeUsers := make(map[string]time.Time)

	for _, user := range allMembers {
		username := user.GetLogin()
		fmt.Printf("Checking activity for user: %s\n", username)

		events, _, err := client.Activity.ListEventsPerformedByUser(ctx, username, false, &github.ListOptions{PerPage: 100})
		if err != nil {
			log.Printf("Error fetching events for %s: %v", username, err)
			continue
		}

		for _, event := range events {
			if event.CreatedAt != nil && event.CreatedAt.Time.After(cutoff) {
				if prevTime, ok := activeUsers[username]; !ok || event.CreatedAt.Time.After(prevTime) {
					activeUsers[username] = event.CreatedAt.Time
				}
			}
		}
	}

	file, err := os.Create("lastactivity.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Println("\n✅ Active users in the last 90 days (with latest activity timestamp):")
	for user, lastActive := range activeUsers {
		line := fmt.Sprintf("%s - %s\n", user, lastActive.Format(time.RFC3339))
		fmt.Print(line)
		file.WriteString(line)
	}
}
