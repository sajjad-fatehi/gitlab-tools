package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sajjad-fatehi/gitlab-tools/internal/bulkmr"
	"github.com/sajjad-fatehi/gitlab-tools/internal/gitlab"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ", ")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	// Try to load .env file (silently fail if not present)
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "bulk-mr":
		bulkMRCommand()
	case "bulk-mr-topic":
		bulkMRTopicCommand()
	case "merge":
		mergeCommand()
	case "topics":
		topicsCommand()
	case "projects":
		projectsCommand()
	case "version":
		fmt.Println("gitlab-tools v1.0.0")
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("GitLab Tools - CLI toolkit for managing GitLab branches and merge requests")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gitlab-tools <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  bulk-mr         Create bulk merge requests across multiple projects")
	fmt.Println("  bulk-mr-topic   Create bulk merge requests for all projects in a topic")
	fmt.Println("  merge           Interactively merge open MRs by target branch and topic")
	fmt.Println("  topics          List all GitLab topics")
	fmt.Println("  projects        List all projects for a specific topic")
	fmt.Println("  version    Show version information")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Run 'gitlab-tools <command> --help' for more information on a command.")
}

func bulkMRCommand() {
	fs := flag.NewFlagSet("bulk-mr", flag.ExitOnError)

	origin := fs.String("origin", "", "Origin (source) branch name (required)")
	target := fs.String("target", "", "Target branch name (required)")
	gitlabURL := fs.String("gitlab-url", os.Getenv("GITLAB_BASE_URL"), "GitLab base URL (default: GITLAB_BASE_URL env)")
	token := fs.String("token", os.Getenv("GITLAB_TOKEN"), "GitLab API token (default: GITLAB_TOKEN env)")
	group := fs.String("group", "", "Default group/namespace prefix (optional)")
	verbose := fs.Bool("verbose", false, "Enable verbose logging")

	var projects arrayFlags
	fs.Var(&projects, "project", "Project path (can be repeated)")

	fs.Usage = func() {
		fmt.Println("Create merge requests from origin branch to target branch across multiple projects")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  gitlab-tools bulk-mr --origin <branch> --target <branch> --project <path> [--project <path>...]")
		fmt.Println()
		fmt.Println("Options:")
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Create MRs from op-stage to op-rc for two projects")
		fmt.Println("  gitlab-tools bulk-mr --origin op-stage --target op-rc \\")
		fmt.Println("    --project group/repo-a --project group/repo-b")
		fmt.Println()
		fmt.Println("  # With group prefix")
		fmt.Println("  gitlab-tools bulk-mr --origin op-stage --target op-rc \\")
		fmt.Println("    --group mygroup --project repo-a --project repo-b")
		fmt.Println()
		fmt.Println("Environment Variables:")
		fmt.Println("  GITLAB_BASE_URL    GitLab instance base URL (e.g., https://gitlab.example.com)")
		fmt.Println("  GITLAB_TOKEN       Personal access token for GitLab API")
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	if *origin == "" {
		fmt.Fprintln(os.Stderr, "Error: --origin is required")
		fs.Usage()
		os.Exit(1)
	}

	if *target == "" {
		fmt.Fprintln(os.Stderr, "Error: --target is required")
		fs.Usage()
		os.Exit(1)
	}

	if len(projects) == 0 {
		fmt.Fprintln(os.Stderr, "Error: at least one --project is required")
		fs.Usage()
		os.Exit(1)
	}

	if *gitlabURL == "" {
		fmt.Fprintln(os.Stderr, "Error: GitLab URL must be provided via --gitlab-url or GITLAB_BASE_URL env")
		fs.Usage()
		os.Exit(1)
	}

	if *token == "" {
		fmt.Fprintln(os.Stderr, "Error: GitLab token must be provided via --token or GITLAB_TOKEN env")
		fs.Usage()
		os.Exit(1)
	}

	processedProjects := make([]string, len(projects))
	for i, project := range projects {
		if *group != "" && !strings.Contains(project, "/") {
			processedProjects[i] = fmt.Sprintf("%s/%s", *group, project)
		} else {
			processedProjects[i] = project
		}
	}

	config := bulkmr.Config{
		OriginBranch: *origin,
		TargetBranch: *target,
		Projects:     processedProjects,
		Verbose:      *verbose,
	}

	if *verbose {
		log.SetFlags(log.Ltime)
	} else {
		log.SetFlags(0)
	}

	fmt.Printf("Processing %d project(s)...\n\n", len(processedProjects))

	client := gitlab.NewClient(*gitlabURL, *token, *verbose)
	service := bulkmr.NewService(client, config)

	results, summary := service.ProcessProjects()

	for _, result := range results {
		printResult(result)
	}

	fmt.Println()
	printSummary(summary)

	if summary.Errors > 0 {
		os.Exit(1)
	}
}

func printResult(result bulkmr.ProjectResult) {
	statusIcon := getStatusIcon(result.Status)
	fmt.Printf("[%s] %s %s\n", result.Project, statusIcon, result.Status)

	if result.Details != "" {
		fmt.Printf("  %s\n", result.Details)
	}

	if result.ErrorMessage != "" {
		fmt.Printf("  Error: %s\n", result.ErrorMessage)
	}

	fmt.Println()
}

func getStatusIcon(status bulkmr.ResultStatus) string {
	switch status {
	case bulkmr.StatusCreated:
		return "✓"
	case bulkmr.StatusSkippedExists:
		return "→"
	case bulkmr.StatusSkippedDraft:
		return "⊘"
	case bulkmr.StatusSkippedBranch:
		return "⚠"
	case bulkmr.StatusSkippedNoChange:
		return "≡"
	case bulkmr.StatusError:
		return "✗"
	default:
		return "?"
	}
}

func printSummary(summary bulkmr.Summary) {
	fmt.Println("Summary:")
	fmt.Printf("  Total projects: %d\n", summary.Total)
	fmt.Printf("  Created: %d\n", summary.Created)
	fmt.Printf("  Skipped (exists): %d\n", summary.SkippedExists)
	fmt.Printf("  Skipped (draft): %d\n", summary.SkippedDraft)
	fmt.Printf("  Skipped (no changes): %d\n", summary.SkippedNoChange)
	fmt.Printf("  Skipped (no branch): %d\n", summary.SkippedBranch)
	fmt.Printf("  Errors: %d\n", summary.Errors)
	fmt.Println()

	if summary.Errors == 0 {
		fmt.Println("✓ Completed successfully")
	} else {
		fmt.Println("✗ Completed with errors")
	}
}

func topicsCommand() {
	fs := flag.NewFlagSet("topics", flag.ExitOnError)

	gitlabURL := fs.String("gitlab-url", os.Getenv("GITLAB_BASE_URL"), "GitLab base URL (default: GITLAB_BASE_URL env)")
	token := fs.String("token", os.Getenv("GITLAB_TOKEN"), "GitLab API token (default: GITLAB_TOKEN env)")
	perPage := fs.Int("per-page", 50, "Number of topics per page")
	page := fs.Int("page", 1, "Page number")
	verbose := fs.Bool("verbose", false, "Enable verbose logging")

	fs.Usage = func() {
		fmt.Println("List all GitLab topics")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  gitlab-tools topics [options]")
		fmt.Println()
		fmt.Println("Options:")
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # List all topics")
		fmt.Println("  gitlab-tools topics")
		fmt.Println()
		fmt.Println("  # List topics with pagination")
		fmt.Println("  gitlab-tools topics --page 2 --per-page 20")
		fmt.Println()
		fmt.Println("Environment Variables:")
		fmt.Println("  GITLAB_BASE_URL    GitLab instance base URL (e.g., https://gitlab.example.com)")
		fmt.Println("  GITLAB_TOKEN       Personal access token for GitLab API")
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	if *gitlabURL == "" {
		fmt.Fprintln(os.Stderr, "Error: GitLab URL must be provided via --gitlab-url or GITLAB_BASE_URL env")
		fs.Usage()
		os.Exit(1)
	}

	if *token == "" {
		fmt.Fprintln(os.Stderr, "Error: GitLab token must be provided via --token or GITLAB_TOKEN env")
		fs.Usage()
		os.Exit(1)
	}

	client := gitlab.NewClient(*gitlabURL, *token, *verbose)
	topics, err := client.ListTopics(*page, *perPage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching topics: %v\n", err)
		os.Exit(1)
	}

	renderTopics(topics)
}

func projectsCommand() {
	fs := flag.NewFlagSet("projects", flag.ExitOnError)

	gitlabURL := fs.String("gitlab-url", os.Getenv("GITLAB_BASE_URL"), "GitLab base URL (default: GITLAB_BASE_URL env)")
	token := fs.String("token", os.Getenv("GITLAB_TOKEN"), "GitLab API token (default: GITLAB_TOKEN env)")
	topic := fs.String("topic", "", "Topic name (required)")
	perPage := fs.Int("per-page", 50, "Number of projects per page")
	page := fs.Int("page", 1, "Page number")
	verbose := fs.Bool("verbose", false, "Enable verbose logging")

	fs.Usage = func() {
		fmt.Println("List all projects for a specific topic")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  gitlab-tools projects --topic <topic-name> [options]")
		fmt.Println()
		fmt.Println("Options:")
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # List all projects for a topic")
		fmt.Println("  gitlab-tools projects --topic backend")
		fmt.Println()
		fmt.Println("  # List projects with pagination")
		fmt.Println("  gitlab-tools projects --topic backend --page 2 --per-page 20")
		fmt.Println()
		fmt.Println("Environment Variables:")
		fmt.Println("  GITLAB_BASE_URL    GitLab instance base URL (e.g., https://gitlab.example.com)")
		fmt.Println("  GITLAB_TOKEN       Personal access token for GitLab API")
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	if *topic == "" {
		fmt.Fprintln(os.Stderr, "Error: --topic is required")
		fs.Usage()
		os.Exit(1)
	}

	if *gitlabURL == "" {
		fmt.Fprintln(os.Stderr, "Error: GitLab URL must be provided via --gitlab-url or GITLAB_BASE_URL env")
		fs.Usage()
		os.Exit(1)
	}

	if *token == "" {
		fmt.Fprintln(os.Stderr, "Error: GitLab token must be provided via --token or GITLAB_TOKEN env")
		fs.Usage()
		os.Exit(1)
	}

	client := gitlab.NewClient(*gitlabURL, *token, *verbose)
	projects, err := client.ListProjectsByTopic(*topic, *page, *perPage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching projects: %v\n", err)
		os.Exit(1)
	}

	renderProjects(*topic, projects)
}

func renderTopics(topics []gitlab.Topic) {
	fmt.Printf("\n📚 \033[1;35mGitLab Topics\033[0m\n\n")

	if len(topics) == 0 {
		fmt.Println("\033[2mNo topics found.\033[0m")
		return
	}

	for i, topic := range topics {
		fmt.Printf("\033[1m%d.\033[0m \033[1;35m%s\033[0m", i+1, topic.Name)

		if topic.Title != "" && topic.Title != topic.Name {
			fmt.Printf(" - %s", topic.Title)
		}

		fmt.Println()

		if topic.TotalProjectsCount > 0 {
			fmt.Printf("   \033[35m📦 %d projects\033[0m\n", topic.TotalProjectsCount)
		}

		if topic.Description != "" && topic.Description != topic.Title {
			desc := topic.Description
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}
			fmt.Printf("   \033[2m%s\033[0m\n", desc)
		}

		fmt.Println()
	}
}

func renderProjects(topicName string, projects []gitlab.Project) {
	fmt.Printf("\n📁 \033[1;35mProjects in topic: %s\033[0m\n\n", topicName)

	if len(projects) == 0 {
		fmt.Println("\033[2mNo projects found for this topic.\033[0m")
		return
	}

	for i, project := range projects {
		fmt.Printf("\033[1m%d.\033[0m \033[1m%s\033[0m\n", i+1, project.Name)
		fmt.Printf("   \033[2m%s\033[0m\n", project.PathWithNamespace)

		if project.Description != "" {
			desc := project.Description
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}
			fmt.Printf("   \033[3;2m%s\033[0m\n", desc)
		}

		if len(project.Topics) > 0 {
			fmt.Print("   ")
			for _, t := range project.Topics {
				if t != topicName {
					fmt.Printf("\033[45m\033[37m %s \033[0m ", t)
				}
			}
			fmt.Println()
		}

		fmt.Printf("   \033[4;32m%s\033[0m\n", project.WebURL)
		fmt.Println()
	}
}

func bulkMRTopicCommand() {
	fs := flag.NewFlagSet("bulk-mr-topic", flag.ExitOnError)

	origin := fs.String("origin", "", "Origin (source) branch name (required)")
	target := fs.String("target", "", "Target branch name (required)")
	topic := fs.String("topic", "", "Topic name (required)")
	gitlabURL := fs.String("gitlab-url", os.Getenv("GITLAB_BASE_URL"), "GitLab base URL (default: GITLAB_BASE_URL env)")
	token := fs.String("token", os.Getenv("GITLAB_TOKEN"), "GitLab API token (default: GITLAB_TOKEN env)")
	perPage := fs.Int("per-page", 100, "Number of projects to fetch per page")
	verbose := fs.Bool("verbose", false, "Enable verbose logging")

	fs.Usage = func() {
		fmt.Println("Create merge requests from origin to target branch for all projects in a topic")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  gitlab-tools bulk-mr-topic --origin <branch> --target <branch> --topic <topic>")
		fmt.Println()
		fmt.Println("Options:")
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Create MRs from op-stage to op-rc for all backend projects")
		fmt.Println("  gitlab-tools bulk-mr-topic --origin op-stage --target op-rc --topic backend")
		fmt.Println()
		fmt.Println("  # With verbose output")
		fmt.Println("  gitlab-tools bulk-mr-topic --origin op-stage --target op-rc --topic frontend --verbose")
		fmt.Println()
		fmt.Println("Environment Variables:")
		fmt.Println("  GITLAB_BASE_URL    GitLab instance base URL (e.g., https://gitlab.example.com)")
		fmt.Println("  GITLAB_TOKEN       Personal access token for GitLab API")
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	if *origin == "" {
		fmt.Fprintln(os.Stderr, "Error: --origin is required")
		fs.Usage()
		os.Exit(1)
	}

	if *target == "" {
		fmt.Fprintln(os.Stderr, "Error: --target is required")
		fs.Usage()
		os.Exit(1)
	}

	if *topic == "" {
		fmt.Fprintln(os.Stderr, "Error: --topic is required")
		fs.Usage()
		os.Exit(1)
	}

	if *gitlabURL == "" {
		fmt.Fprintln(os.Stderr, "Error: GitLab URL must be provided via --gitlab-url or GITLAB_BASE_URL env")
		fs.Usage()
		os.Exit(1)
	}

	if *token == "" {
		fmt.Fprintln(os.Stderr, "Error: GitLab token must be provided via --token or GITLAB_TOKEN env")
		fs.Usage()
		os.Exit(1)
	}

	if *verbose {
		log.SetFlags(log.Ltime)
	} else {
		log.SetFlags(0)
	}

	client := gitlab.NewClient(*gitlabURL, *token, *verbose)

	fmt.Printf("Fetching projects for topic: \033[1;35m%s\033[0m\n\n", *topic)

	// Fetch all projects for the topic (handle pagination)
	var allProjects []gitlab.Project
	page := 1
	for {
		projects, err := client.ListProjectsByTopic(*topic, page, *perPage)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching projects: %v\n", err)
			os.Exit(1)
		}

		if len(projects) == 0 {
			break
		}

		allProjects = append(allProjects, projects...)

		if len(projects) < *perPage {
			break
		}

		page++
	}

	if len(allProjects) == 0 {
		fmt.Printf("No projects found for topic: %s\n", *topic)
		os.Exit(0)
	}

	// Convert projects to paths
	projectPaths := make([]string, len(allProjects))
	for i, project := range allProjects {
		projectPaths[i] = project.PathWithNamespace
	}

	fmt.Printf("Found %d project(s) in topic \033[1;35m%s\033[0m\n\n", len(projectPaths), *topic)

	config := bulkmr.Config{
		OriginBranch: *origin,
		TargetBranch: *target,
		Projects:     projectPaths,
		Verbose:      *verbose,
	}

	service := bulkmr.NewService(client, config)
	results, summary := service.ProcessProjects()

	for _, result := range results {
		printResult(result)
	}

	fmt.Println()
	printSummary(summary)

	if summary.Errors > 0 {
		os.Exit(1)
	}
}

func mergeCommand() {
	mergeCmd := flag.NewFlagSet("merge", flag.ExitOnError)
	target := mergeCmd.String("target", "", "Target branch to merge into (required)")
	topic := mergeCmd.String("topic", "", "Topic to filter projects (required)")

	mergeCmd.Parse(os.Args[2:])

	if *target == "" || *topic == "" {
		fmt.Println("\033[31mError: both --target and --topic are required\033[0m")
		mergeCmd.PrintDefaults()
		os.Exit(1)
	}

	baseURL := os.Getenv("GITLAB_BASE_URL")
	token := os.Getenv("GITLAB_TOKEN")

	if baseURL == "" || token == "" {
		log.Fatal("GITLAB_BASE_URL and GITLAB_TOKEN environment variables are required")
	}

	client := gitlab.NewClient(baseURL, token, false)

	// Get all projects for the topic
	fmt.Printf("\033[36m📦 Fetching projects for topic: %s\033[0m\n", *topic)
	projects, err := client.ListProjectsByTopic(*topic, 1, 100)
	if err != nil {
		log.Fatalf("Failed to fetch projects: %v", err)
	}

	if len(projects) == 0 {
		fmt.Printf("\033[33m⚠️  No projects found for topic: %s\033[0m\n", *topic)
		return
	}

	fmt.Printf("\033[32m✓ Found %d projects\033[0m\n\n", len(projects))

	scanner := bufio.NewScanner(os.Stdin)
	mergedCount := 0
	skippedCount := 0
	errorCount := 0

	// Process each project
	for _, project := range projects {
		// Get open merge requests for this project targeting the specified branch
		mrs, err := client.ListOpenMergeRequestsByTarget(project.ID, *target)
		if err != nil {
			fmt.Printf("\033[31m✗ Error fetching MRs for %s: %v\033[0m\n", project.PathWithNamespace, err)
			errorCount++
			continue
		}

		// Filter out draft MRs
		nonDraftMRs := []gitlab.MergeRequest{}
		for _, mr := range mrs {
			if !mr.IsDraft() {
				nonDraftMRs = append(nonDraftMRs, mr)
			}
		}

		if len(nonDraftMRs) == 0 {
			continue
		}

		// Display and prompt for each MR
		for _, mr := range nonDraftMRs {
			fmt.Printf("\033[36m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\033[0m\n")
			fmt.Printf("\033[1;36mProject:\033[0m %s\n", project.PathWithNamespace)
			fmt.Printf("\033[1;36mMR Title:\033[0m %s\n", mr.Title)
			fmt.Printf("\033[1;36mBranches:\033[0m %s → %s\n", mr.SourceBranch, mr.TargetBranch)
			fmt.Printf("\033[1;36mURL:\033[0m %s\n", mr.WebURL)
			fmt.Printf("\033[36m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\033[0m\n")
			fmt.Print("\033[1;33mMerge this MR? (y/n): \033[0m")

			if !scanner.Scan() {
				break
			}

			response := strings.ToLower(strings.TrimSpace(scanner.Text()))

			if response == "y" || response == "yes" {
				_, err := client.AcceptMergeRequest(project.ID, mr.IID)
				if err != nil {
					fmt.Printf("\033[31m✗ Failed to merge: %v\033[0m\n\n", err)
					errorCount++
				} else {
					fmt.Printf("\033[32m✓ Successfully merged!\033[0m\n\n")
					mergedCount++
				}
			} else {
				fmt.Printf("\033[33m⊘ Skipped\033[0m\n\n")
				skippedCount++
			}
		}
	}

	// Print summary
	fmt.Printf("\033[36m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\033[0m\n")
	fmt.Printf("\033[1;36m📊 Summary\033[0m\n")
	fmt.Printf("\033[36m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\033[0m\n")
	fmt.Printf("\033[32m✓ Merged:  %d\033[0m\n", mergedCount)
	fmt.Printf("\033[33m⊘ Skipped: %d\033[0m\n", skippedCount)
	if errorCount > 0 {
		fmt.Printf("\033[31m✗ Errors:  %d\033[0m\n", errorCount)
	}
	fmt.Printf("\033[36m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\033[0m\n")

	if errorCount > 0 {
		os.Exit(1)
	}
}
