package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	// "strings"
	"time"

	"github.com/gorilla/mux" // Import Gorilla Mux router
    "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	connectString = "mongodb+srv://fazlulkarim362:PNl5egJWL6LkIbEB@cluster0.t1ag31c.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
	dbName        = "FirstGolangDatabase"
	colName       = "crud-operation"
)

type MongoField struct {
	BookName   string `json:"bookName"`
	AuthorName string `json:"authorName"`
	Rating     int    `json:"rating"`
}

var collection *mongo.Collection

func init() {
	clientOption := options.Client().ApplyURI(connectString)

	client, err := mongo.Connect(context.TODO(), clientOption)
	if err != nil {
		panic(err)
	}
	fmt.Println("Mongo Connection Successful")
	collection = client.Database(dbName).Collection(colName)
}

func insertData(w http.ResponseWriter, r *http.Request) {
	var data MongoField
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert data into MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	_, err = collection.InsertOne(ctx, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func updateBooks(w http.ResponseWriter, r *http.Request) {
    // Get the bookName from the URL parameters
    vars := mux.Vars(r)
    bookName := vars["bookName"]

    // Parse the request body to get the updated data
    var data MongoField
    err := json.NewDecoder(r.Body).Decode(&data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Update data in MongoDB
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
    defer cancel()

    filter := bson.M{"bookname": bookName} // Filter to find the documents
    update := bson.M{"$set": bson.M{"authorName": data.AuthorName, "rating": data.Rating}} // Update fields
    result, err := collection.UpdateMany(ctx, filter, update)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    response := map[string]interface{}{
        "message":         "Data updated successfully",
        "matchedCount":    result.MatchedCount,   // Number of documents matched by the filter
        "modifiedCount":   result.ModifiedCount,  // Number of documents modified
        "updatedBookName": bookName,
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

func getAllBooks(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
    defer cancel()

    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)

    var books []MongoField
    for cursor.Next(ctx) {
        var book MongoField
        if err := cursor.Decode(&book); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        books = append(books, book)
    }
    if err := cursor.Err(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(books)
}
func deleteBook(w http.ResponseWriter, r *http.Request) {
    var data MongoField
    err := json.NewDecoder(r.Body).Decode(&data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Delete data from MongoDB
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
    defer cancel()

    filter := bson.M{"bookname": data.BookName} // Filter to find the document
    result, err := collection.DeleteOne(ctx, filter)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if result.DeletedCount == 0 {
        http.Error(w, "No document found with the given bookName", http.StatusNotFound)
        return
    }

    response := map[string]interface{}{
        "message":       "Data deleted successfully",
        "deletedCount":  result.DeletedCount,
        "deletedBook":   data.BookName,
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to the server!")
	})


	r.HandleFunc("/books", insertData).Methods("POST")
    
	r.HandleFunc("/books/update/{bookName}", updateBooks).Methods("PUT")


	r.HandleFunc("/books", getAllBooks).Methods("GET")

	r.HandleFunc("/books",deleteBook).Methods("DELETE")

	fmt.Println("Server started on port 5000")
	log.Fatal(http.ListenAndServe(":5000", r))
}
