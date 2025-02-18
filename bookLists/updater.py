import json
import os
import requests
import time

def getFileName():
    while True:
        bookFile = input("Please input the file name for a list of books:")
        if os.path.exists(bookFile):
            return loadBooks(bookFile)
        else:
            print("Hey, thats not a file, what are you doing")
            print("Try again big guy")

def loadBooks(file):
    if os.path.exists(file):
        with open(file, "r") as f:
            return json.load(f)
    else:
        return []
def saveBooks(updatedBooks):
    with open("populatedBooks.json", "w") as f:
        json.dump(updatedBooks, f, indent=4)

def updateBooks(books):
    if books == []:
       print("no books :(")
       return
    for idx, book in enumerate(books, start=1):
        print(idx)
        missingKeys = missingData(book)
        if missingKeys:
            print(missingKeys)
            updatedBook = updateBook(book, missingKeys)
            books[idx-1] = updatedBook
        time.sleep(1)
    return books

def missingData(book):
    keys = ["title", "author", "genre", "isbn", "published_date"]
    for k in keys:
        if book.get(k) is None or book[k] == "" or book[k] == "Unknown" or book[k] == "unknown":
            continue
        keys.remove(k)
    return keys

def handleWeirdCase(book, vol_inf, isbn, genre, published_date, ret_title):
    #got title != retrieved title -> either we have case mismatch, one title is a subset or its the straight up wrong book
    title = book["title"].lower()
    ret = ret_title.lower()
    if title == ret:
        book["isbn"] = isbn
        book["genre"] = genre
        book["published_date"] = published_date
        print("case mismatch False alarm!")
        return book
    if title in ret or ret in title:
        book["isbn"] = isbn
        book["genre"] = genre
        book["published_date"] = published_date
        print("subset problem, false alarm!")
        return book
    print("Got the wrong book! Asked for:")
    print(title)
    print("Returned:")
    print(ret)
    time.sleep(1)
    with open("missingTitles.txt", "a") as missing:
        missing.write(book["title"])
    return book


def updateBook(book, missingKeys):
    title = book["title"]
    author = book["author"]
    api_url = "https://www.googleapis.com/books/v1/volumes"
    params = {
        "q": f'intitle:{title} inauthor:{author}',
        "maxResults": 1,  # Get only the most relevant book
        "printType": "books"
    }
    print("Making call to api")
    print(title)
    print(author)
    response = requests.get(api_url, params=params)
    book_data = response.json()
    if "items" in book_data:
        volume_info = book_data["items"][0]["volumeInfo"]
        isbn = next((id["identifier"] for id in volume_info.get("industryIdentifiers", []) if id["type"] == "ISBN_13"), None)
        genre = volume_info.get("categories", ["Unknown"])[0]
        retrieved_title = volume_info.get("title", "How did this happen")
        published_date = volume_info.get("publishedDate", "Unknown")
        if retrieved_title != title:
            print('Something weird is going on, got title {} but requested title {}', retrieved_title, title)
            handleWeirdCase(book, volume_info, isbn, genre, published_date, retrieved_title)
            return book
        book["isbn"] = isbn
        book["genre"] = genre
        book["published_date"] = published_date
    else:
        print("Failed to get book with title:", title)
        print(response)
    print(book)
    return book

def callApi():
    api_url = "https://www.googleapis.com/books/v1/volumes"
    title = "Vision of the Annointed"
    author = "Thomas Sowell"
    params = {
        "q": f'intitle:{title} inauthor:{author}',
        "maxResults": 1,  # Get only the most relevant book
        "printType": "books"
    }

    response = requests.get(api_url, params=params)
    book_data = response.json()
    if "items" in book_data:
        volume_info = book_data["items"][0]["volumeInfo"]
        isbn = next((id["identifier"] for id in volume_info.get("industryIdentifiers", []) if id["type"] == "ISBN_13"), None)
        genre = volume_info.get("categories", ["Unknown"])[0]
        retrieved_title = volume_info.get("title")
        published_date = volume_info.get("publishedDate", "Unknown")
    if retrieved_title != title:
        print("Something weird is going on here!")
        print(retrieved_title)
        print(title)
    print("Isbn:{}", isbn)
    time.sleep(1)
    print("Genre: {}", genre)
    time.sleep(1)
    print("Date: {}", published_date)
    time.sleep(1)
    return book_data


def main():
    callApi()
    fileName = getFileName()
    updatedBooks = updateBooks(fileName)
    saveBooks(updatedBooks)
    print("Books Updated Successfully :)")    

if __name__ == "__main__":
    main()
