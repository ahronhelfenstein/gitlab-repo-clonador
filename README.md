# GitLab Repo Clonador

This Go program fetches groups, subgroups, and projects from a GitLab server using its API. It clones all projects into a specified directory. If a project already exists locally, it skips cloning and notifies the user.

## Features

- Fetch GitLab groups, subgroups (up to 2 levels), and their projects.
- Clone projects into a specified local directory.
- Skip cloning if the project already exists.

---

## Prerequisites

1. Go installed ([Download Go](https://golang.org/dl/))
2. A GitLab Access Token with API access.
3. Git installed on your system.

---

## Usage

1. **Build the program**:

   ```bash
   go build -o gitlab-repo-clonador main.go
   ```


2. **Run the program**:

Use the following flags to pass the required parameters:

- `-accessToken`: Your GitLab API access token.
- `-baseDir`: The local directory where projects will be cloned.
- `-gitlabBaseURL`: The GitLab hostname (e.g., `gitlab.com`).
- `-groupIf`: The group ID to fetch subgroups and projects for.

Example:
```bash
./gitlab-cloner -accessToken="YOUR_ACCESS_TOKEN" -baseDir="/path/to/clone" -gitlabBaseURL="gitlab.com" -groupId="123456"
```

## Example Output
```bash
Cloning into 'project-name-1'...
✅ Project: https://git.example.com/group/project-name-1.git cloned into folder: /path/to/clone/group/project-name-1...

Cloning into 'project-name-2'...
✅ Project: https://git.example.com/group/project-name-2.git cloned into folder: /path/to/clone/group/project-name-2...


🔍 Skipping cloning for project: https://gitlab/my-group/my-subgroup/project3. Directory already exists at: /path/to/clone/my-group/my-subgroup
🎉 All Git clones completed successfully! 🎉
```