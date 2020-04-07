FROM amazonlinux
COPY main main
ENTRYPOINT ["./main"]
CMD ["3"]
