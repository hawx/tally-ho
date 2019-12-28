# tally-ho

A micropub speaking blog.

Relevant specs:

- [Micropub](https://www.w3.org/TR/micropub/)
- [Webmention](https://www.w3.org/TR/webmention/)
- [IndieAuth](https://www.w3.org/TR/indieauth/)

Possibly up-to-date feature list:

- IndieAuth:
  * [x] Authentication in header
  * [x] Authentication in body

- Config:
  * [x] `q=config`
  * [x] `q=source`
  * [x] `q=syndicateTo`
  * [ ] Get `q` options

- Posting:
  * [x] Create with `application/x-www-form-urlencoded`
  * [x] Create with `application/json`
  * [ ] Create with `multipart/form-data`
  * [x] Update with `application/json` (although does not yet require `update` scope)
  * [x] Upload to media endpoint
  * [ ] Delete
  * [ ] Undelete

- Syndication:
  * Twitter:
    * [x] Create (only like: `like-of`, or post: `content`)
    * [ ] Retrieve likes
    * [ ] Retrieve replies
    * [ ] Retrieve retweets
  * Flickr
    * [ ] Upload photo
    * [ ] Retreive likes
    * [ ] Retrieve comments
  * GitHub
    * [ ] Create issue
    * [ ] Create comment
    * [ ] Retrieve reactions

- Webmentions:
  * [x] Receive webmentions for posts
  * [ ] Send webmentions on create
  * [ ] Send webmentions on update
  * [ ] Send webmentions on delete

- Display:
  * List:
    * [x] All
    * [x] Combine likes
    * [ ] Pagination
    * [ ] By kind
  * Entry:
    * [x] Notes
    * [x] Posts
    * [x] Photos
    * [ ] Videos
    * [x] Likes
    * [ ] Replies
    * [ ] RSVP
