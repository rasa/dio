package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	rest "github.com/emicklei/go-restful"
)

func main() {
	// Create the storage directories on disk
	err := os.MkdirAll(filepath.Join(STORAGEDIR, "files"), os.ModeDir|0755)
	if err != nil {
		log.Printf("Something went wrong when creating the files dir: %v\n", err.Error())
		return
	}
	err = os.MkdirAll(filepath.Join(STORAGEDIR, "meta"), os.ModeDir|0755)
	if err != nil {
		log.Printf("Something went wrong when creating the meta dir: %v\n", err.Error())
		return
	}

	// Create and start the API server
	ws := new(rest.WebService)
	ws.Filter(rest.NoBrowserCacheFilter)
	ws.Route(ws.POST("/branch_create").Consumes("application/x-www-form-urlencoded").To(branchCreate))
	ws.Route(ws.GET("/branch_history").To(branchHistory))
	ws.Route(ws.GET("/branch_list").To(branchList))
	ws.Route(ws.POST("/branch_remove").Consumes("application/x-www-form-urlencoded").To(branchRemove))
	ws.Route(ws.POST("/branch_revert").Consumes("application/x-www-form-urlencoded").To(branchRevert))
	ws.Route(ws.PUT("/db_upload").To(dbUpload))
	ws.Route(ws.GET("/db_download").To(dbDownload))
	ws.Route(ws.GET("/db_list").To(dbList))
	rest.Add(ws)
	http.ListenAndServe(":8080", nil)
}

// Creates a new branch for a database.
// Can be tested with: curl -d database=a.db -d branch=master -d commit=xxx http://localhost:8080/branch_create
func branchCreate(r *rest.Request, w *rest.Response) {
	// Retrieve the database and branch names
	err := r.Request.ParseForm()
	if err != nil {
		w.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	dbName := r.Request.FormValue("database")
	branchName := r.Request.FormValue("branch")
	commit := r.Request.FormValue("commit")

	// Sanity check the inputs
	if dbName == "" || branchName == "" || commit == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO: Validate the database and branch names

	// Ensure the requested database is in our system
	if !dbExists(dbName) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Load the existing branch heads from disk
	branches, err := getBranches(dbName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Ensure the new branch name doesn't already exist in the database
	if _, ok := branches[branchName]; ok {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// Ensure the requested commit exists in the database history
	// This is important, because if we didn't do it then people could supply any commit ID.  Even one's belonging
	// to completely unrelated databases, which they shouldn't have access to.
	// By ensuring the requested commit is already part of the existing database history, we solve that problem.
	commitExists := false
	for _, ID := range branches {
		// Because there seems to be no valid way of adding this condition in the loop declaration?
		// Do I just need more coffee? :D
		if commitExists == true {
			continue
		}

		// Check if the head commit of the branch matches the requested commit
		if ID == commit {
			commitExists = true
			continue
		}

		// It didn't, so walk the commit history looking for the commit there
		c, err := getCommit(ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for c.Parent != "" {
			if c.ID == commit {
				// This commit in the history matches the requested commit, so we're good
				commitExists = true
			}
			c, err = getCommit(c.Parent)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}

	// The commit wasn't found, so don't create the new branch
	if commitExists != true {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Create the new branch
	branches[branchName] = commit
	err = storeBranches(dbName, branches)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Returns the history for a branch.
// Can be tested with: curl -H "Database: a.db" -H "Branch: master" http://localhost:8080/branch_history
func branchHistory(r *rest.Request, w *rest.Response) {
	// Retrieve the database and branch names
	dbName := r.Request.Header.Get("Database")
	branchName := r.Request.Header.Get("Branch")

	// TODO: Validate the database and branch names

	// Sanity check the inputs
	if dbName == "" || branchName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Ensure the requested database is in our system
	if !dbExists(dbName) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Load the existing branch heads from disk
	branches, err := getBranches(dbName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Ensure the requested branch exists in the database
	id, ok := branches[branchName]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Walk the commit history, assembling it into something useful
	var history []commit
	c, err := getCommit(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	history = append(history, c)
	for c.Parent != "" {
		c, err = getCommit(c.Parent)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		history = append(history, c)
	}
	w.WriteAsJson(history)
}

// Returns the list of branch heads for a database.
// Can be tested with: curl -H "Database: a.db" http://localhost:8080/branch_list
func branchList(r *rest.Request, w *rest.Response) {
	// Retrieve the database name
	dbName := r.Request.Header.Get("Database")

	// TODO: Validate the database name

	// Sanity check the input
	if dbName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Ensure the requested database is in our system
	if !dbExists(dbName) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Load the existing branch heads from disk
	branches, err := getBranches(dbName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return the list of branch heads
	w.WriteAsJson(branches)
}

// Removes a branch from a database.
// Can be tested with: curl -d database=a.db -d branch=master http://localhost:8080/branch_remove
func branchRemove(r *rest.Request, w *rest.Response) {
	// Retrieve the database and branch names
	err := r.Request.ParseForm()
	if err != nil {
		w.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	dbName := r.Request.FormValue("database")
	branchName := r.Request.FormValue("branch")

	// Sanity check the inputs
	if dbName == "" || branchName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO: Validate the database and branch names

	// Ensure the requested database is in our system
	if !dbExists(dbName) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Load the existing branch heads from disk
	branches, err := getBranches(dbName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Ensure the branch exists in the database
	if _, ok := branches[branchName]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Remove the branch
	delete(branches, branchName)
	err = storeBranches(dbName, branches)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Reverts a branch to a previous commit.
// Can be tested with: curl -d database=a.db -d branch=master -d commit=xxx http://localhost:8080/branch_revert
func branchRevert(r *rest.Request, w *rest.Response) {
	// Retrieve the database and branch names
	err := r.Request.ParseForm()
	if err != nil {
		w.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	dbName := r.Request.FormValue("database")
	branchName := r.Request.FormValue("branch")
	commit := r.Request.FormValue("commit")

	// Sanity check the inputs
	if dbName == "" || branchName == "" || commit == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO: Validate the database and branch names

	// Ensure the requested database is in our system
	if !dbExists(dbName) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Load the existing branch heads from disk
	branches, err := getBranches(dbName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Ensure the branch exists in the database
	id, ok := branches[branchName]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// * Ensure the requested commit exists in the branch history *

	// If the head commit of the branch already matches the requested commit, there's nothing to change
	if id == commit {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// It didn't, so walk the branch history looking for the commit there
	commitExists := false
	c, err := getCommit(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for c.Parent != "" {
		if c.ID == commit {
			// This commit in the branch history matches the requested commit, so we're good to proceed
			commitExists = true
		}
		c, err = getCommit(c.Parent)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// The commit wasn't found, so don't update the branch
	if commitExists != true {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Update the branch
	branches[branchName] = commit
	err = storeBranches(dbName, branches)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Upload a database.
// Can be tested with: curl -T a.db -H "Name: a.db" -w \%{response_code} -D headers.out http://localhost:8080/db_upload
func dbUpload(r *rest.Request, w *rest.Response) {
	// Retrieve the database and branch names
	dbName := r.Request.Header.Get("Name")
	branchName := r.Request.Header.Get("Branch")

	// TODO: Validate the database and branch names

	// Sanity check the inputs
	if dbName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// If no branch name was given, use the default for the database
	if branchName == "" {
		branchName = getDefaultBranchName(dbName)
	}

	// Read the database into a buffer
	var buf bytes.Buffer
	buf.ReadFrom(r.Request.Body)
	sha := sha256.Sum256(buf.Bytes())

	// Create a dbTree entry for the individual database file
	var e dbTreeEntry
	e.AType = DATABASE
	e.Sha256 = hex.EncodeToString(sha[:])
	e.Name = dbName
	e.Last_Modified = time.Now()
	e.Size = buf.Len()

	// Create a dbTree structure for the database entry
	var t dbTree
	t.Entries = append(t.Entries, e)
	t.ID = createDBTreeID(t.Entries)

	// Construct a commit structure pointing to the tree
	var c commit
	c.AuthorEmail = "justin@postgresql.org" // TODO: Author and Committer info should come from the client, so we
	c.AuthorName = "Justin Clift"           // TODO  hard code these for now.  Proper auth will need adding later
	c.Timestamp = time.Now()                // TODO: If provided from the client, use a timestamp for the db
	c.Tree = t.ID

	// Check if the database already exists
	var err error
	var branches map[string]string
	needDefBranch := false
	if dbExists(dbName) {
		// Load the existing branchHeads from disk
		branches, err = getBranches(dbName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Check if the desired branch already exists.  If it does, use its head commit as the parent for
		// our new uploads commit
		if id, ok := branches[branchName]; ok {
			c.Parent = id
		}
	} else {
		// No existing branches, so this will be the first
		branches = make(map[string]string)

		// We'll need to create the default branch value for the database later on too
		needDefBranch = true
	}

	// Update the branch with the commit for this new database upload
	c.ID = createCommitID(c)
	branches[branchName] = c.ID
	err = storeDatabase(buf.Bytes())
	if err != nil {
		log.Printf("Error when writing database '%s' to disk: %v\n", dbName, err.Error())

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Write the tree to disk
	err = storeTree(t)
	if err != nil {
		log.Printf("Something went wrong when storing the tree file: %v\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Write the commit to disk
	err = storeCommit(c)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Write the updated branch heads to disk
	err = storeBranches(dbName, branches)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Write the default branch name to disk
	if needDefBranch {
		err = storeDefaultBranchName(dbName, branchName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Log the upload
	log.Printf("Database uploaded.  Name: '%s', size: %d bytes, branch: '%s'\n", dbName, buf.Len(),
		branchName)

	// Send a 201 "Created" response, along with the location of the URL for working with the (new) database
	w.AddHeader("Location", "/"+dbName)
	w.WriteHeader(http.StatusCreated)
}

// Download a database
func dbDownload(r *rest.Request, w *rest.Response) {
	log.Println("dbDownload() called")
}

// Get a list of databases
func dbList(r *rest.Request, w *rest.Response) {
	log.Println("dbList() called")
}