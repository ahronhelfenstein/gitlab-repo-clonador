package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
)

type Subgroup struct {
	ID       int    `json:"id"`
	Path     string `json:"path"`
	FullPath string `json:"full_path"`
}

type Project struct {
	ID       int    `json:"id"`
	HttpsUrl string `json:"http_url_to_repo"`
	Path     string `json:"path"`
}

var (
	accessToken   string
	baseDir       string
	gitlabBaseURL string
	groupId       string
)

func main() {
	flag.StringVar(&accessToken, "accessToken", "", "GitLab Access Token (required)")
	flag.StringVar(&baseDir, "baseDir", "", "Base directory to clone projects into (required)")
	flag.StringVar(&gitlabBaseURL, "gitlabBaseURL", "", "GitLab Hostname (e.g., gitlab.com) (required)")
	flag.StringVar(&groupId, "groupId", "", "Group to fetch for its subgroups and projects (required)")
	flag.Parse()

	// Validate required inputs
	if accessToken == "" || baseDir == "" || gitlabBaseURL == "" || groupId == "" {
		log.Fatalf("All flags -accessToken, -baseDir, -gitlabBaseURL, and -groupId are required.")
	}

	// Construct full GitLab API URL
	gitlabBaseURL = fmt.Sprintf("https://%s/api/v4", gitlabBaseURL)

	mainGroupInfo := fetchSubGroups(getGroupApiURL(groupId))[0]
	listProjectsAndClone(strconv.Itoa(mainGroupInfo.ID), mainGroupInfo.FullPath)

	subgroups := fetchSubGroups(getSubGroupdApiURL(strconv.Itoa(mainGroupInfo.ID)))
	//basePath := path.Join(baseDir, groupId)

	// first subGroup level
	for _, subgroup := range subgroups {
		listProjectsAndClone(strconv.Itoa(subgroup.ID), subgroup.FullPath)

		// goes one level deeper
		childSubgroups := fetchSubGroups(getSubGroupdApiURL(strconv.Itoa(subgroup.ID)))

		for _, childSubgroup := range childSubgroups {
			listProjectsAndClone(strconv.Itoa(childSubgroup.ID), childSubgroup.FullPath)
		}
	}
	// Print final message when all cloning is done
	log.Println("🎉 All Git clones completed successfully! 🎉")
}

func getGroupApiURL(groupParam string) string {
	return fmt.Sprintf("%s/groups/%s/", gitlabBaseURL, groupParam)
}

func getSubGroupdApiURL(groupParam string) string {
	return getGroupApiURL(groupParam) + "subgroups/"
}

func fetchSubGroups(url string) []Subgroup {
	var subgroups []Subgroup

	// Make the request
	resp, err := request(url)
	if err != nil {
		log.Fatalf("Error fetching subgroups: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatalf("Failed to close response body: %v", err)
		}
	}()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	if err := json.Unmarshal(body, &subgroups); err != nil {
		var subgroup Subgroup
		if err := json.Unmarshal(body, &subgroup); err != nil {
			log.Fatalf("Error unmarshaling response body: %v", err)
		}
		// Wrap the single object in a slice
		subgroups = append(subgroups, subgroup)
	}

	return subgroups
}

func listProjectsAndClone(groupId string, subGroupPath string) {
	projectsUrl := fmt.Sprintf("%s/groups/%s/projects?archived=false&per_page=1000&with_shared=false", gitlabBaseURL, groupId)
	resp, err := request(projectsUrl)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	// Read and print the raw response body
	body, err := io.ReadAll(resp.Body)

	var projects []Project

	err = json.Unmarshal(body, &projects)

	for _, project := range projects {
		err := cloneGitProject(project.HttpsUrl, subGroupPath, project.Path)
		if err != nil {
			log.Fatalf("Error occurred: %v\n", err)
		}
	}

}

func cloneGitProject(gitProjectUrl string, fullPath string, projectPath string) error {
	cloneDir := path.Join(baseDir, fullPath)
	if _, err := os.Stat(path.Join(cloneDir, projectPath)); err == nil {
		log.Printf("🔍 Skipping cloning for project: %s. Directory already exists at: %s\n", gitProjectUrl, cloneDir)
		return nil
	}

	// Ensure the target directory exists, create it if it doesn't
	if err := os.MkdirAll(cloneDir, os.ModePerm); err != nil {
		log.Printf("failed to create directory %s: %v", cloneDir, err)
	}

	cmd := exec.Command("git", "clone", gitProjectUrl)
	cmd.Dir = cloneDir // Set working directory where to run the git command
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to clone project %s: %v", gitProjectUrl, err)
	}

	log.Printf(" ✅ Project: %s cloned into folder: %s...\n", gitProjectUrl, cloneDir)

	return nil
}

func request(url string) (*http.Response, error) {

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("failed to create request: %v", err)
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", accessToken)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failed to send request: %v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("unexpected response: %s, body: %s", resp.Status, string(body))
		return nil, fmt.Errorf("unexpected response: %s", resp.Status)
	}

	return resp, nil
}
