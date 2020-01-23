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
  * [x] Get `q` options
  * [x] Micropub `q=config`
  * [x] Micropub `q=media-endpoint`
  * [x] Micropub `q=source`
  * [x] Micropub `q=syndicate-to`
  * [x] Media `q=last`

- Posting:
  * [x] Create with `application/x-www-form-urlencoded`
  * [x] Create with `application/json`
  * [x] Create with `multipart/form-data`
    * [x] Store photo/audio/video as if they had been sent via the media endpoint
  * [x] Update with `application/json`
    * [ ] Require `update` scope for requests
  * [x] Upload to media endpoint
  * [x] Delete
    * [x] `410 Gone` entry
    * [x] Remove from listing
    * [ ] Remove from grouped likes
  * [x] Undelete
  * [ ] `mp-slug`
  * [ ] `post-status`
  * [ ] [indiebookclub](https://indiebookclub.biz/documentation#micropub)

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
    * [ ] By category
  * Entry:
    * [x] Notes
    * [x] Posts
    * [x] Photos
    * [ ] Videos
    * [x] Likes
    * [ ] Replies
    * [x] RSVP
    * [x] Checkin (kinda)

- Feeds:
  * [ ] RSS
  * [ ] Atom
  * [ ] Jsonfeed (<https://jsonfeed.org/>)
  * [ ] WebSub (<https://www.w3.org/TR/websub/>)
