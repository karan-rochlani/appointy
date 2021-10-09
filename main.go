package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var client *mongo.Client

type Post struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID    string             `json:"userid,omitempty" bson:"userid,omitempty"`
	Caption   string             `json:"caption,omitempty" bson:"caption,omitempty"`
	ImageURL  string             `json:"imageurl,omitempty" bson:"imageurl,omitempty"`
	Timestamp time.Time          `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
}
type User struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name,omitempty" bson:"name,omitempty"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty"`
	Password string             `json:"password,omitempty" bson:"password,omitempty"`
}

func main() {
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	http.HandleFunc("/U", root)
	http.HandleFunc("/users", CreateUserEndpoint)
	http.HandleFunc("/users/", GetUserEndpoint)
	http.HandleFunc("/posts", CreatePostEndpoint)
	http.HandleFunc("/posts/", GetPostEndpoint)
	http.HandleFunc("/posts/users/", GetPostUserEndpoint)
	http.ListenAndServe(":12345", nil)
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Appointy Instagram API\n\n19BCE0383\nKaran Rochlani\nVellore Institute of Technology")
	fmt.Println("Displaying Root.")
}

//Create a User using post request and also display all the users using get request
func CreateUserEndpoint(response http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		response.Header().Set("content-type", "application/json")
		var user User
		_ = json.NewDecoder(request.Body).Decode(&user)
		user.Password = protect(user.Password)
		collection := client.Database("appointy").Collection("users")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		result, _ := collection.InsertOne(ctx, user)
		json.NewEncoder(response).Encode(result)
	} else {
		response.Header().Set("content-type", "application/json")
		var users []User
		collection := client.Database("appointy").Collection("users")
		ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
		cursor, err := collection.Find(ctx, bson.M{})
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var user User
			cursor.Decode(&user)
			users = append(users, user)
		}
		if err := cursor.Err(); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		json.NewEncoder(response).Encode(users)
	}
}

//Passwords securely stored
func protect(pwd string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	return string(hash)
}

//Get a user using id
func GetUserEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	id := strings.TrimPrefix(request.URL.Path, "/users/")
	objID, _ := primitive.ObjectIDFromHex(id)
	var user User
	collection := client.Database("appointy").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, User{ID: objID}).Decode(&user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(user)
}

//Create a Post using post request and also display all the posts using get request
func CreatePostEndpoint(response http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		response.Header().Set("content-type", "application/json")
		var post Post
		_ = json.NewDecoder(request.Body).Decode(&post)
		collection := client.Database("appointy").Collection("posts")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		post.Timestamp = time.Now()
		result, _ := collection.InsertOne(ctx, post)
		json.NewEncoder(response).Encode(result)
	} else {
		response.Header().Set("content-type", "application/json")
		var posts []Post
		collection := client.Database("appointy").Collection("posts")
		ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
		cursor, err := collection.Find(ctx, bson.M{})
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var post Post
			cursor.Decode(&post)
			posts = append(posts, post)
		}
		if err := cursor.Err(); err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		json.NewEncoder(response).Encode(posts)
	}
}

//Get a post using id
func GetPostEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	id := strings.TrimPrefix(request.URL.Path, "/posts/")
	objID, _ := primitive.ObjectIDFromHex(id)
	var post Post
	collection := client.Database("appointy").Collection("posts")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, Post{ID: objID}).Decode(&post)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(post)
}

//List all posts of a user
func GetPostUserEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	id := strings.TrimPrefix(request.URL.Path, "/posts/users/")
	var posts []Post
	collection := client.Database("appointy").Collection("posts")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	cursor, err := collection.Find(ctx, Post{UserID: id})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var post Post
		cursor.Decode(&post)
		posts = append(posts, post)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(posts)
}
