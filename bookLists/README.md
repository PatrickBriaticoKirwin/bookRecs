Here are a few more scripts I used to populate the list of books I had.
append.go appends an existing json of books with more from the next page of recs.
dedup de-duplicated a list of title originally. I remember there was a list of recommendations (or rather more specifically, mentions) that had duplicate entries. in its current state it outputs a list of titles that were not updated from calling the googleBooks api.
Lister listed books. I honestly have no idea why it exists. I don't remember using it.
Updater is probably the only good script in this batch of files because you actually input a file name.
Takes in a json of books, and fetches more data about it, then adds that to a new file. That file only contains the books where the api call actually got the book we wanted.

