package languages

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aretext/aretext/syntax/parser"
)

func TestDockerfileParseFunc(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []TokenWithText
	}{
		{
			name: "comments and parser directives",
			text: `# syntax=docker/dockerfile:1
  # comment
RUN echo '# not a dockerfile comment'`,
			expected: []TokenWithText{
				{Text: "# syntax=docker/dockerfile:1\n", Role: parser.TokenRoleComment},
				{Text: "# comment\n", Role: parser.TokenRoleComment},
				{Text: "RUN", Role: parser.TokenRoleKeyword},
				{Text: "'# not a dockerfile comment'", Role: parser.TokenRoleString},
			},
		},
		{
			name: "line leading whitespace before instruction",
			text: `   	FROM alpine
	COPY . /app`,
			expected: []TokenWithText{
				{Text: "FROM", Role: parser.TokenRoleKeyword},
				{Text: "COPY", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "from with stage alias",
			text: `FROM --platform=$BUILDPLATFORM golang:1.22 AS build
FROM scratch`,
			expected: []TokenWithText{
				{Text: "FROM", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "$BUILDPLATFORM", Role: bashTokenRoleVariable},
				{Text: "AS", Role: parser.TokenRoleKeyword},
				{Text: "FROM", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "add instruction",
			text: `ADD file /dest`,
			expected: []TokenWithText{
				{Text: "ADD", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "arg instruction",
			text: `ARG VERSION=latest`,
			expected: []TokenWithText{
				{Text: "ARG", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "cmd instruction",
			text: `CMD echo hello`,
			expected: []TokenWithText{
				{Text: "CMD", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "copy instruction",
			text: `COPY . /app`,
			expected: []TokenWithText{
				{Text: "COPY", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "entrypoint instruction",
			text: `ENTRYPOINT echo hello`,
			expected: []TokenWithText{
				{Text: "ENTRYPOINT", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "env instruction",
			text: `ENV APP_HOME=/app`,
			expected: []TokenWithText{
				{Text: "ENV", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
			},
		},
		{
			name: "expose instruction",
			text: `EXPOSE 8080`,
			expected: []TokenWithText{
				{Text: "EXPOSE", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "from instruction",
			text: `FROM alpine`,
			expected: []TokenWithText{
				{Text: "FROM", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "healthcheck instruction",
			text: `HEALTHCHECK CMD curl -f http://localhost/ || exit 1`,
			expected: []TokenWithText{
				{Text: "HEALTHCHECK", Role: parser.TokenRoleKeyword},
				{Text: "CMD", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "label instruction",
			text: `LABEL version="1.0"`,
			expected: []TokenWithText{
				{Text: "LABEL", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: `"1.0"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "maintainer instruction",
			text: `MAINTAINER user@example.com`,
			expected: []TokenWithText{
				{Text: "MAINTAINER", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "onbuild instruction",
			text: `ONBUILD COPY . /app`,
			expected: []TokenWithText{
				{Text: "ONBUILD", Role: parser.TokenRoleKeyword},
				{Text: "COPY", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "run instruction",
			text: `RUN echo hello`,
			expected: []TokenWithText{
				{Text: "RUN", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "shell instruction",
			text: `SHELL ["/bin/bash", "-c"]`,
			expected: []TokenWithText{
				{Text: "SHELL", Role: parser.TokenRoleKeyword},
				{Text: `"/bin/bash"`, Role: parser.TokenRoleString},
				{Text: `"-c"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "stopsignal instruction",
			text: `STOPSIGNAL SIGTERM`,
			expected: []TokenWithText{
				{Text: "STOPSIGNAL", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "user instruction",
			text: `USER app:app`,
			expected: []TokenWithText{
				{Text: "USER", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "volume instruction",
			text: `VOLUME /data`,
			expected: []TokenWithText{
				{Text: "VOLUME", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "workdir instruction",
			text: `WORKDIR /app`,
			expected: []TokenWithText{
				{Text: "WORKDIR", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "instructions are case insensitive",
			text: `from alpine as base
rUn echo hello`,
			expected: []TokenWithText{
				{Text: "from", Role: parser.TokenRoleKeyword},
				{Text: "as", Role: parser.TokenRoleKeyword},
				{Text: "rUn", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "exec form cmd",
			text: `CMD ["echo", "hello world"]`,
			expected: []TokenWithText{
				{Text: "CMD", Role: parser.TokenRoleKeyword},
				{Text: `"echo"`, Role: parser.TokenRoleString},
				{Text: `"hello world"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "shell form cmd",
			text: `CMD echo "hello world"`,
			expected: []TokenWithText{
				{Text: "CMD", Role: parser.TokenRoleKeyword},
				{Text: `"hello world"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "json array forms",
			text: `RUN ["echo", "hello"]
ENTRYPOINT ["top", "-b"]`,
			expected: []TokenWithText{
				{Text: "RUN", Role: parser.TokenRoleKeyword},
				{Text: `"echo"`, Role: parser.TokenRoleString},
				{Text: `"hello"`, Role: parser.TokenRoleString},
				{Text: "ENTRYPOINT", Role: parser.TokenRoleKeyword},
				{Text: `"top"`, Role: parser.TokenRoleString},
				{Text: `"-b"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "top level quoted values",
			text: `LABEL description="a \"quoted\" value"
ENV APP_HOME="/usr/local/app"`,
			expected: []TokenWithText{
				{Text: "LABEL", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: `"a \"quoted\" value"`, Role: parser.TokenRoleString},
				{Text: "ENV", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: `"/usr/local/app"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "line continuation keeps next line out of top level",
			text: `RUN echo hello \
    && echo FROM
COPY . /app`,
			expected: []TokenWithText{
				{Text: "RUN", Role: parser.TokenRoleKeyword},
				{Text: "&&", Role: parser.TokenRoleOperator},
				{Text: "COPY", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "comment line inside continued run command",
			text: `RUN echo hello \
# comment removed before shell runs
world
COPY . /app`,
			expected: []TokenWithText{
				{Text: "RUN", Role: parser.TokenRoleKeyword},
				{Text: "# comment removed before shell runs\n", Role: parser.TokenRoleComment},
				{Text: "COPY", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "escape directive changes line continuation character",
			text: "# escape=`\nRUN echo hello `\n    && echo FROM\nCOPY . /app",
			expected: []TokenWithText{
				{Text: "# escape=`\n", Role: parser.TokenRoleComment},
				{Text: "RUN", Role: parser.TokenRoleKeyword},
				{Text: "COPY", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "full multi stage dockerfile",
			text: `# syntax=docker/dockerfile:1

FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN ["go", "test", "./..."]

FROM gcr.io/distroless/base-debian12
COPY --from=build /src/app /app
USER nonroot:nonroot
ENTRYPOINT ["/app"]`,
			expected: []TokenWithText{
				{Text: "# syntax=docker/dockerfile:1\n", Role: parser.TokenRoleComment},
				{Text: "FROM", Role: parser.TokenRoleKeyword},
				{Text: "AS", Role: parser.TokenRoleKeyword},
				{Text: "WORKDIR", Role: parser.TokenRoleKeyword},
				{Text: "COPY", Role: parser.TokenRoleKeyword},
				{Text: "RUN", Role: parser.TokenRoleKeyword},
				{Text: "COPY", Role: parser.TokenRoleKeyword},
				{Text: "RUN", Role: parser.TokenRoleKeyword},
				{Text: `"go"`, Role: parser.TokenRoleString},
				{Text: `"test"`, Role: parser.TokenRoleString},
				{Text: `"./..."`, Role: parser.TokenRoleString},
				{Text: "FROM", Role: parser.TokenRoleKeyword},
				{Text: "COPY", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "USER", Role: parser.TokenRoleKeyword},
				{Text: "ENTRYPOINT", Role: parser.TokenRoleKeyword},
				{Text: `"/app"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "full service dockerfile",
			text: `FROM alpine:3.20
LABEL org.opencontainers.image.title="example"
ENV APP_HOME=/srv/app
WORKDIR $APP_HOME
RUN apk add --no-cache curl \
    && adduser -D app
EXPOSE 8080/tcp
HEALTHCHECK --interval=30s CMD curl -f http://localhost:8080/health || exit 1
USER app
CMD ["./server", "--listen", ":8080"]`,
			expected: []TokenWithText{
				{Text: "FROM", Role: parser.TokenRoleKeyword},
				{Text: "LABEL", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: `"example"`, Role: parser.TokenRoleString},
				{Text: "ENV", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "WORKDIR", Role: parser.TokenRoleKeyword},
				{Text: "$APP_HOME", Role: bashTokenRoleVariable},
				{Text: "RUN", Role: parser.TokenRoleKeyword},
				{Text: "&&", Role: parser.TokenRoleOperator},
				{Text: "EXPOSE", Role: parser.TokenRoleKeyword},
				{Text: "HEALTHCHECK", Role: parser.TokenRoleKeyword},
				{Text: "=", Role: parser.TokenRoleOperator},
				{Text: "CMD", Role: parser.TokenRoleKeyword},
				{Text: "||", Role: parser.TokenRoleOperator},
				{Text: "USER", Role: parser.TokenRoleKeyword},
				{Text: "CMD", Role: parser.TokenRoleKeyword},
				{Text: `"./server"`, Role: parser.TokenRoleString},
				{Text: `"--listen"`, Role: parser.TokenRoleString},
				{Text: `":8080"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "full trigger dockerfile",
			text: `ARG BASE=alpine
FROM ${BASE}
MAINTAINER user@example.com
ONBUILD ADD . /src
ONBUILD RUN make -C /src
VOLUME ["/cache", "/data"]
STOPSIGNAL SIGQUIT
SHELL ["/bin/sh", "-c"]`,
			expected: []TokenWithText{
				{Text: "ARG", Role: parser.TokenRoleKeyword},
				{Text: "FROM", Role: parser.TokenRoleKeyword},
				{Text: "MAINTAINER", Role: parser.TokenRoleKeyword},
				{Text: "ONBUILD", Role: parser.TokenRoleKeyword},
				{Text: "ADD", Role: parser.TokenRoleKeyword},
				{Text: "ONBUILD", Role: parser.TokenRoleKeyword},
				{Text: "RUN", Role: parser.TokenRoleKeyword},
				{Text: "VOLUME", Role: parser.TokenRoleKeyword},
				{Text: `"/cache"`, Role: parser.TokenRoleString},
				{Text: `"/data"`, Role: parser.TokenRoleString},
				{Text: "STOPSIGNAL", Role: parser.TokenRoleKeyword},
				{Text: "SHELL", Role: parser.TokenRoleKeyword},
				{Text: `"/bin/sh"`, Role: parser.TokenRoleString},
				{Text: `"-c"`, Role: parser.TokenRoleString},
			},
		},
		{
			name: "heredoc body is not top level dockerfile syntax",
			text: `RUN <<EOF
FROM busybox
# not a dockerfile comment
EOF
COPY . /app`,
			expected: []TokenWithText{
				{Text: "RUN", Role: parser.TokenRoleKeyword},
				{Text: "COPY", Role: parser.TokenRoleKeyword},
			},
		},
		{
			name: "instruction keywords are not recognized inside words",
			text: `BEFORE alpine
RUNFROM busybox
COPYRIGHT file`,
			expected: []TokenWithText{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := ParseTokensWithText(DockerfileParseFunc(), tc.text)
			assert.Equal(t, tc.expected, tokens)
		})
	}
}
