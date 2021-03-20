# tally-ho

A micropub speaking blog.

Relevant specs:

- [Micropub](https://www.w3.org/TR/micropub/)
- [Webmention](https://www.w3.org/TR/webmention/)
- [IndieAuth](https://www.w3.org/TR/indieauth/)
- [WebSub](https://www.w3.org/TR/websub/)

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
    * [x] Require `update` scope for requests
  * [x] Upload to media endpoint
  * [x] Delete
    * [x] `410 Gone` entry
    * [x] Remove from listing
    * [x] Remove from grouped likes
  * [x] Undelete
  * [ ] `mp-slug`
  * [ ] `post-status`

- Syndication:
  * Twitter
    * Create
        * [x] Notes
        * [ ] Posts
        * [x] Photos
        * [ ] Videos
        * [x] Likes
        * [x] Replies
        * [x] Reposts (url only)
    * [ ] Retrieve likes
    * [ ] Retrieve replies
    * [ ] Retrieve retweets
  * Flickr
    * Create
      * [ ] Photos
      * [ ] Videos
      * [x] Likes
      * [x] Replies
    * [ ] Retreive likes
    * [ ] Retrieve comments
  * GitHub
    * [x] Likes
    * [ ] Create issue
    * [ ] Create comment
    * [ ] Retrieve reactions

- Webmentions:
  * [x] Receive webmentions for posts
  * [x] Send webmentions on create
  * [x] Send webmentions on update
  * [x] Send webmentions on delete
  * [x] Send webmentions on undelete

- Display:
  * List:
    * [x] All
    * [x] Combine likes
    * [x] Pagination
    * [x] By kind
    * [x] By category
  * Entry:
    * [x] Notes
    * [x] Posts
    * [x] Photos
    * [ ] Videos
    * [x] Likes
    * [x] Replies
    * [x] Bookmarks
    * [x] RSVPs
    * [x] Checkins (kinda)
    * [x] Reposts
    * [x] [indiebookclub](https://indiebookclub.biz/)
    * [x] [teacup](https://teacup.p3k.io/)

- Feeds:
  * [x] RSS
  * [x] Atom
  * [x] Jsonfeed (<https://jsonfeed.org/>)
  * [x] WebSub
    * [x] On create
    * [x] On update
    * [x] On delete
    * [x] On undelete
