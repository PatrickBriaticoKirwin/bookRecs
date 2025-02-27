This is a simple Go service to fetch the book recommendations from a db I created. 
To collect the initial lists of book recommendations, I found pages where the authors recommendations appeared either in static websites made by them, collected by others, or on goodreads.
Then I hit the Google Books api to get more info about the books, as the non-goodreads lists had missing data (Genre, date, isbn).

Included in the bookLists directory are some of the scripts I used to seed my local db and collect the above data. Before using source control I had several iterations of these scripts, so names and their current functions are not complete.
There should also be a form here to add more recommendations, as the people I have so far are still alive and suggesting books for me to purchase and never read (I never learned how to read).

Overall, this probably could have taken me a day but I ran out of tokens pretty often.

If I were to do this again I'd probably set up the go service to be not just one main function.
