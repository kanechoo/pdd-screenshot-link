FROM gocv/opencv:4.10.0-ubuntu-22.04
WORKDIR /app
RUN wget https://go.dev/dl/go1.22.7.linux-amd64.tar.gz
RUN tar -xzvf go1.22.7.linux-amd64.tar.gz > /dev/null 2>&1
ENV GOROOT=/app/go
ENV PATH=$PATH:$GOROOT/bin
# copy the source code to the container
COPY . .
ENV HTTP_PROXY=guest:xxx@45.45.218.90:3128
ENV HTTPS_PROXY=guest:xxx@45.45.218.90:3128
# build the source code
RUN go build -o app .
RUN apt update && apt install -y gnupg
RUN wget -O - https://notesalexp.org/debian/alexp_key.asc | apt-key add -
RUN echo "deb https://notesalexp.org/tesseract-ocr5/jammy/ jammy main" | tee /etc/apt/sources.list.d/notesalexp.list > /dev/null
RUN apt update
RUN apt-cache policy tesseract-ocr
# Add tesseract-ocr
RUN apt install -y tesseract-ocr && apt install -y curl
# create a directory name tessdata
RUN mkdir -p tessdata
# download tesseract traineddata to the tessdata directory
RUN curl -L -o tessdata/eng.traineddata https://github.com/tesseract-ocr/tessdata/blob/main/eng.traineddata?raw=true
ENV TESSDATA_DIR=/app/tessdata
ENV HTTP_PROXY=""
ENV HTTPS_PROXY=""
#expose the port
EXPOSE 8188
# run the application
CMD ["./app"]


