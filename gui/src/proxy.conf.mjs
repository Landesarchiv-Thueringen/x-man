export default {
  "/api": {
    target: "http://localhost:3000",
    secure: false,
    // Forward connection closing to the server. This is relevant for
    // server-sent events, since the server keeps the connection open
    // indefinitely otherwise.
    onProxyReq: (proxyRes, req, res) => {
      res.on("close", () => proxyRes.destroy());
    },
  },
};
