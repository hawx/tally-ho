#!/usr/bin/env sh

set -e

# Usage: create_posts.sh URL TOKEN
#
#  - URL is where tally-ho is running, e.g. http://localhost:8080
#  - TOKEN is a valid auth token

URL=$1
TOKEN=$2

# simple post
curl -XPOST $URL/-/micropub -H "Authorization: Bearer $2" -H "Content-Type: application/x-www-form-urlencoded" \
     -d 'h=entry' \
     -d 'content=hello+world' \
     -d 'category[]=foo&category[]=bar'

# html post
curl -XPOST $URL/-/micropub -H "Authorization: Bearer $2" -H "Content-Type: application/json" \
     -d '{
  "type": ["h-entry"],
  "properties": {
    "name": ["Itching: h-event to iCal converter"],
    "content": [
      {"html": "Now that I have been <a href=\"http://localhost:9999\">creating a list of events</a> on my site using <a href=\"http://localhost:9999\">p3k</a>, it would be great if I could get a more calendar-like view of that list..."}
    ],
    "category": [
      "indieweb", "p3k"
    ]
  }
}'

# photo post

# bookmark
curl -XPOST $URL/-/micropub -H "Authorization: Bearer $2" -H "Content-Type: application/x-www-form-urlencoded" \
     -d 'h=entry' \
     -d 'name=Interesting' \
     -d 'content=I+was+thinking+about+this...' \
     -d 'bookmark-of=http://localhost:9999'

# rsvp
curl -XPOST $URL/-/micropub -H "Authorization: Bearer $2" -H "Content-Type: application/x-www-form-urlencoded" \
     -d 'h=entry' \
     -d 'name=An+event' \
     -d 'rsvp=yes'

# reply
curl -XPOST $URL/-/micropub -H "Authorization: Bearer $2" -H "Content-Type: application/x-www-form-urlencoded" \
     -d 'h=entry' \
     -d 'content=Here+are+my+thoughts' \
     -d 'in-reply-to=http://localhost:9999' \
     -d 'category=foo'

# like
curl -XPOST $URL/-/micropub -H "Authorization: Bearer $2" -H "Content-Type: application/x-www-form-urlencoded" \
     -d 'h=entry' \
     -d 'like-of=http://localhost:9999'

# repost
curl -XPOST $URL/-/micropub -H "Authorization: Bearer $2" -H "Content-Type: application/x-www-form-urlencoded" \
     -d 'h=entry' \
     -d 'repost-of=http://localhost:9999'

# drink
curl -XPOST $URL/-/micropub -H "Authorization: Bearer $2" -H "Content-Type: application/json" \
     -d '{
  "type": ["h-entry"],
  "properties": {
    "drank": [
      {
        "type": ["h-food"],
        "properties": {
          "name": ["Tea"]
        }
      }
    ]
  }
}'
