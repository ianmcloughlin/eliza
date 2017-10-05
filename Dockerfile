# add a blank docker image
FROM scratch
WORKDIR /app
# Now just add the binary and text files
ADD . /app/
ENTRYPOINT ["./eliza"]