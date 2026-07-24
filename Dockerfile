FROM gcr.io/distroless/static-debian12:nonroot

# Release wiring will copy the GoReleaser Linux binary into this context.
COPY vizb /vizb

ENTRYPOINT ["/vizb"]
CMD ["serve", "--host", "0.0.0.0", "--port", "8080"]
