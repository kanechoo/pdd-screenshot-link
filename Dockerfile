FROM gocv/opencv:4.10.0-ubuntu-22.04

WORKDIR /build
COPY . .
ENV GOROOT=/build/go
ENV PATH=$PATH:$GOROOT/bin
ENV HTTP_PROXY=guest:xxx@45.45.218.90:3128
ENV HTTPS_PROXY=guest:xxx@45.45.218.90:3128
RUN wget https://go.dev/dl/go1.22.7.linux-amd64.tar.gz && \
    tar -xzvf go1.22.7.linux-amd64.tar.gz > /dev/null 2>&1 && \
    go build -o app .
WORKDIR /app
RUN mkdir tessdata && \
    apt update && \
    apt install -y gnupg && \
    wget -O - https://notesalexp.org/debian/alexp_key.asc | apt-key add - && \
    echo "deb https://notesalexp.org/tesseract-ocr5/jammy/ jammy main" | tee /etc/apt/sources.list.d/notesalexp.list > /dev/null && \
    apt update && \
    apt install -y tesseract-ocr && \
    apt install -y tesseract-ocr-chi-sim && \
    apt install -y curl && \
    curl -L -o tessdata/eng.traineddata https://github.com/tesseract-ocr/tessdata/blob/main/eng.traineddata?raw=true && \
    cp /build/app /app && \
    rm -rf /build
ENV TESSDATA_DIR=/app/tessdata
ENV HTTP_PROXY=""
ENV HTTPS_PROXY=""
EXPOSE 8188
CMD ["./app"]


