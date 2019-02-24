FROM golang:alpine

RUN apk add --no-cache \
			shadow \
			gcc \
			musl-dev \
			curl \
			gzip \
			zip \
			git

RUN useradd -m -s /bin/sh builder

USER builder

WORKDIR /home/builder

RUN curl -L https://github.com/tehmoon/cryptocli/releases/download/latest/cryptocli-linux-amd64.gz | gunzip > cryptocli && chmod +x cryptocli

# expire in 10 mins
CMD ["sleep", "600"]
