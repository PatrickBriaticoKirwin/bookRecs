import json

with open("books.json", "r", encoding="utf-8") as f:
    books = json.load(f)

with open("another_book_list.json", "r", encoding="utf-8") as f:
    otherBooks = json.load(f)

authors = {book["author"]:book for book in books}

# Merge details into the full list
for book in otherBooks:
    author = book["author"]
    if author in marcAuthors:
        print("Overlapped Author")
        print(author)
        print("First Book")
        print(authors[author]["title"])
        print("The other book")
        print(book["title"])



