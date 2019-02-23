FROM golang:alpine

RUN apk add --no-cache \
			shadow \
			gcc \
			musl-dev \
			zip \
			git

RUN useradd -m -s /bin/sh builder

USER builder

WORKDIR /home/builder

# expire in 10 mins
CMD ["sleep", "600"]
