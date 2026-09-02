try {
  const status = rs.status();
  if (status.ok !== 1) {
    quit(1);
  }
} catch (error) {
  const result = rs.initiate({
    _id: "rs0",
    members: [{ _id: 0, host: "mongo:27017" }],
  });
  if (result.ok !== 1 && result.codeName !== "AlreadyInitialized") {
    printjson(result);
    quit(1);
  }
}
