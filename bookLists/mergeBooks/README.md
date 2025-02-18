Here are a few scripts as they currently exist on my local machine. 
dbDump dumped a populated list of books into my local db.
standardizeDate standardized the dates so the books with incomplete dates didn't cause a problem with the schema
merge was originally for merging lists that had overlapping books, but now just checks if there is an overlap of authors between two lists.
This was because for whatever reason, two lists had the same book but didn't standardize the same way which made the dumping a mess.
