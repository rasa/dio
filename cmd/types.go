package cmd

import "time"

type branchEntries struct {
	Branch  string
	Entries []commitEntry
}

type branchEntry struct {
	Commit      string `json:"commit"`
	CommitCount int    `json:"commit_count"`
	Description string `json:"description"`
}

type commitEntry struct {
	AuthorEmail    string    `json:"author_email"`
	AuthorName     string    `json:"author_name"`
	CommitterEmail string    `json:"committer_email"`
	CommitterName  string    `json:"committer_name"`
	ID             string    `json:"id"`
	Message        string    `json:"message"`
	Parent         string    `json:"parent"`
	Timestamp      time.Time `json:"timestamp"`
	Tree           dbTree    `json:"tree"`
}

type CommitList struct {
	Commits []commitEntry `json:"commits"`
}

type dbListEntry struct {
	CommitID     string `json:"commit_id"`
	DefBranch    string `json:"default_branch"`
	LastModified string `json:"last_modified"`
	Licence      string `json:"licence"`
	Name         string `json:"name"`
	Public       bool   `json:"public"`
	RepoModified string `json:"repo_modified"`
	SHA256       string `json:"sha256"`
	Size         int    `json:"size"`
	Type         string `json:"type"`
	URL          string `json:"url"`
}

type dbTreeEntryType string

const (
	TREE     dbTreeEntryType = "tree"
	DATABASE                 = "db"
	LICENCE                  = "licence"
)

type dbTree struct {
	ID      string        `json:"id"`
	Entries []dbTreeEntry `json:"entries"`
}
type dbTreeEntry struct {
	AType        dbTreeEntryType `json:"type"`
	LastModified time.Time       `json:"last_modified"`
	Licence      string          `json:"licence"`
	Name         string          `json:"name"`
	Sha256       string          `json:"sha256"`
	Size         int             `json:"size"`
}

type errorInfo struct {
	Condition string   `json:"error_condition"`
	Data      []string `json:"data"`
}

type licenceEntry struct {
	FullName string `json:"full_name"`
	SHA256   string `json:"sha256"`
	Type     string `json:"type"`
	URL      string `json:"url"`
}

type tagType string

const (
	SIMPLE    tagType = "simple"
	ANNOTATED         = "annotated"
)

type tagEntry struct {
	Commit      string    `json:"commit"`
	Date        time.Time `json:"date"`
	Message     string    `json:"message"`
	TagType     tagType   `json:"type"`
	TaggerEmail string    `json:"email"`
	TaggerName  string    `json:"name"`
}
