# medisite

The source for [medidew.name](https://medidew.name), my website serving as a blog, project portfolio, and hub for links to other socials and self-hosted services. It is a small Go web server that renders plain HTML on the server and runs on my own hardware behind an nginx reverse proxy.

The site has three sections:

- **Home**: an introduction to me and to what the site is for.
- **Blog**: posts written in Markdown, listed newest first, with a simple full-text search.
- **Portfolio**: a list of projects, each with a description, tags, and a link.

The design is deliberately plain, as explained on the home page. There is no client-side JavaScript. Blog posts are Markdown files in `content/posts/`, each starting with a YAML frontmatter block. Each page is an `html/template` file in `templates/` that fills in a `content` block inside the shared layout in `base.html.tmpl`.

## Dependencies

- **Go** (1.22+)
- **[goldmark](https://github.com/yuin/goldmark)**
- **[yaml.v3](https://gopkg.in/yaml.v3)**
- **[zap](https://github.com/uber-go/zap)**
- **[lumberjack](https://github.com/natefinch/lumberjack)**
- **systemd**

The web server is designed to run on Linux distributions, but is only tested on Ubuntu Server.

### Installing and redeploying

```bash
sudo ./deploy/install.sh
```

This builds the binary, creates the `medisite` user and the directories above if needed, copies in the binary, config, templates, and static files, installs and enables the systemd unit, and restarts the service. It is safe to re-run for every redeploy.

`install.sh` only copies `content/` on the very first install. After that it leaves the server's copy alone, so posts and portfolio entries can be edited directly on the server without a redeploy overwriting them.

```bash
sudo ./deploy/update.sh             # replace content/ with the repo's copy
sudo ./deploy/update.sh --rollback  # restore the previous content/
```

`update.sh` is for when the repository, not the server, is the source of truth for posts and portfolio entries. It replaces only `content/`, keeping the previous copy alongside as `content.prev`, then restarts the service. It requires an existing install; everything else (binary, templates, static files, config, systemd unit) is handled by `install.sh`.