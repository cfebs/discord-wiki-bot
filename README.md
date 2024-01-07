# discord-wiki-bot

Syncs a list of markdown documents to discord channel messages.

Useful for rules, faqs, wikis dialogs.

## Usage

1. Create 1 git repo per discord server.
1. Create a directory file heirarchy like so:

```
<channel_id>/
    1-<post_title>.md
    2-<post_title_2>.md
    ...
<channel_id>/
    1-<post_title>.md
    2-<post_title_2>.md
    ...
```

1. Run the sync
```
discord-wiki-bot -serverid <id> -basedir <some_dir>
```

## How it works

* Loop through each channel C.
* Read first N messages in channel C that are posted by the bot user.
* Loop through each .md file M.
* If an existing message exists edit it, otherwise create it.

## TODO

* [ ] Cleanup when files are deleted
* [ ] Some tests
* [ ] Document GH action

