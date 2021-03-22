# tally-ho

A micropub speaking blog.


## usage

This is an almost all-in-one solution for running an "IndieWeb" blog. To get
running you will need to:

1. Run `go install hawx.me/code/tally-ho@latest`, or clone this repo and run `go build`

1. Make a directory to put media files in (`tally-ho` will write files to this
   directory but will not serve them)

1. Create a config file, by default it looks for a file called `config.toml`

    ```
    me = "https://john.example.com/"
    name = "John Doe"
    title = "John Doe's blog"
    description = "My great blog"

    # the URL tally-ho will be accessed from
    baseURL = "http://blog.john.example.com/"
    # the URL the media directory will be accessed from
    mediaURL = "http://media.john.example.com/"

    # each of these blocks can be left out if you don't want to use them
    [twitter]
    consumerKey = "..."
    consumerSecret = "..."
    accessToken = "..."
    accessTokenSecret = "..."

    [flickr]
    consumerKey = "..."
    consumerSecret = "..."
    accessToken = "..."
    accessTokenSecret = "..."

    [github]
    clientID = "..."
    clientSecret = "..."
    accessToken = "..."
    ```

1. Copy the [`web`](web) directory somewhere

1. Edit the contents of
   [`web/templates/foooter.gotmpl`](web/templates/footer.gotmpl) to not have my
   details, and potentially change the css or templates to be more your style

Then you are ready to run it:

```
$ tally-ho
    --config $PATH_TO_CONFIG_FILE
    --web $PATH_TO_WEB_DIR
    --media-dir $PATH_TO_MEDIA_DIR
    --db ./db.sqlite
```

It will be listening on <http://localhost:8080>, this can be changed by passing
`--port` or `--socket`. If run as a systemd service then it will detect a
corresponding `.socket` definition.

To get webmentions for social media posts I recommend setting up
<https://brid.gy/>, as `tally-ho` only allows syndicating to
Twitter/Flickr/GitHub and not gathering responses (yet).


## features

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
        * [x] Posts
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
    * Likes
      * [x] Repos
      * [ ] Issues
      * [ ] Comments
    * [x] Create issue (in-reply-to repo)
    * [x] Create comment (in-reply-to issue)
    * [ ] Retrieve reactions
    * [ ] Retrieve comments

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

Relevant specs:

- [Micropub](https://www.w3.org/TR/micropub/)
- [Webmention](https://www.w3.org/TR/webmention/)
- [IndieAuth](https://www.w3.org/TR/indieauth/)
- [WebSub](https://www.w3.org/TR/websub/)
