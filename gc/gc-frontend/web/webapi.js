import SwaggerClient from "swagger-client";
import { objToPipeline } from "./types";

const clientPromise = new SwaggerClient({
  // url: "http://75.16.17.237:80/www/swagger/server.swagger.json",
  url: "http://192.168.143.238:12101/swagger/server.swagger.json",
  usePromise: true
  // authorizations: {}
});

// function getAuthorization() {
//   return {
//     clientAuthorizations: {
//       api_key: new SwaggerClient.ApiKeyAuthorization(
//         "Authorization",
//         localStorage.getItem("token"),
//         "header"
//       )
//     }
//   };
// }

export async function getPipelines() {
  let client = await clientPromise;
  return (
    (await client.Computer.FetchPipelineList({ body: {} })).obj.pipelines || []
  );
}

export async function getPipeline(path, name) {
  let client = await clientPromise;
  let resp = await client.Computer.FetchPipeline({ body: { path, name } });
  return objToPipeline(resp.obj);
}

export async function addPipeline(name, path) {
  let client = await clientPromise;
  await client.Computer.CreatePipeline({ body: { name, path } });
}

export async function updatePipeline(pipeline) {
  let client = await clientPromise;
  console.log("client.Computer", client, client.Computer);
  await client.Computer.UpdatePipeline({ body: pipeline });
}

export async function deletePipeline(path, name) {
  let client = await clientPromise;
  await client.Computer.DeletePipeline({ body: { name, path } });
}

export async function checkPipeline(path, name) {
  let client = await clientPromise;
  await client.Computer.CheckPipeline({ body: { name, path } });
}

export async function getTasks() {
  let client = await clientPromise;
  let resp = await client.Computer.FetchTaskList({ body: {} });
  return resp.obj.tasks || [];
}

export async function getTask(path, name) {
  let client = await clientPromise;
  let resp = await client.Computer.FetchTask({ body: { path, name } });
  return resp.obj;
}

export async function addTask(name, path) {
  let client = await clientPromise;
  await client.Computer.CreateTask({ body: { name, path } });
}

export async function updateTask(task, append_) {
  let client = await clientPromise;
  await client.Computer.UpdateTask({ body: { task, append_ } });
}
export async function recalculateEnabledTasks(path) {
  let client = await clientPromise;
  await client.Computer.RecalculateEnabledTasks({ body: { path } });
}
export async function fetchTestResult(path,name) {
  let client = await clientPromise;
  let resp = await client.Computer.FetchTestResult({ body: { path, name } });
  return resp.obj
}
export async function fetchRunTime(path,name) {
  let client = await clientPromise;
  let resp = await client.Computer.FetchRunTime({ body: { path, name } });
  return resp.obj
}
export async function debugTask(path, name) {
  let client = await clientPromise;
  let resp = await client.Computer.DebugTask({ body: { path, name } });
  return resp.obj.result;
}

export async function deleteTask(path, name) {
  let client = await clientPromise;
  await client.Computer.DeleteTask({ body: { name, path } });
}

export async function duplicatePipeline(path, name, new_path, new_name) {
  let client = await clientPromise;
  await client.Computer.DuplicatePipeline({
    body: { path, name, new_path, new_name }
  });
}

const clientDataPromise = new SwaggerClient({
  url: "http://192.168.143.238:12101/swagger/data_manager.swagger.json",
  usePromise: true
});

export async function listTableMetadata(path, brief_resp) {
  let client = await clientDataPromise;
  let resp = await client.DataManager.TableMetadataList({
    body: { path, brief_resp }
  });
  return resp.obj.tableMetadata || [];
}

export async function getTableMetadata(path, name) {
  let client = await clientDataPromise;
  let resp = await client.DataManager.FetchTableMetadata({
    body: { path, name }
  });
  return resp.obj;
}
export async function deleteTableMetadata(path, name) {
  let client = await clientDataPromise;
  await client.DataManager.DeleteTableMetadata({ body: { name, path } });
}
export async function upsertTableMetadata(tableMetadata) {
  let client = await clientDataPromise;
  await client.DataManager.UpsertTableMetadata({
    body: { table: tableMetadata, removeData: true }
  });
}

export async function listDatabaseTable(path, name) {
  let client = await clientDataPromise;
  let resp = await client.DataManager.FetchDatabaseTableList({
    body: { name, path }
  });
  return resp.obj.databaseTable || [];
}

export async function getDatabaseTable(dbName) {
  let client = await clientDataPromise;
  let resp = await client.DataManager.FetchDatabaseTable({ body: { dbName } });
  return resp.obj;
}

export async function upsertDatabaseTable(databaseTable) {
  let client = await clientDataPromise;
  await client.DataManager.UpsertDatabaseTable({ body: databaseTable });
}

export async function fetchDataFileList(dbName) {
  let client = await clientDataPromise;
  let resp = await client.DataManager.FetchDataFileList({ body: { dbName } });
  return resp.obj.fileNames || [];
}

export async function analyzeDataFile(dbName, fileName) {
  let client = await clientDataPromise;
  let resp = await client.DataManager.AnalyzeDataFile({
    body: { dbName, fileName }
  });
  return (resp.obj.success ? "成功：" : "失败：") + resp.obj.message;
}

// export async function login(name, password) {
//   let client = await clientPromise;
//   let resp = await client.Computer.Login({ body: { name, password } });
//   return resp.obj.token;
// }

export async function fetchDBConnections() {
  let client = await clientPromise;
  let resp = await client.Computer.OracleMsgList({
    body: {}
  });
  return resp.obj.oraclemsgs || [];
}

export async function addDBConnection(
  name,
  user,
  passwd,
  host,
  port,
  instance,
  dbname
) {
  let client = await clientPromise;
  let resp = await client.Computer.InsertOracleMsg({
    body: {
      name,
      usrname: user,
      password: passwd,
      ipadd: host,
      port,
      instance,
      databasename: dbname
    }
  });
  return resp.obj.success;
}

export async function removeDBConnection(name) {
  let client = await clientPromise;
  let resp = await client.Computer.DeleteOracleMsg({
    body: { name }
  });
  return resp.obj.success;
}

export async function fetchDBGroups(conn) {
  let client = await clientDataPromise;
  console.log(client);
  let resp = await client.DataManager.FetchTableGroups({
    body: { conn_name: conn }
  });
  return resp.obj.groups || [];
}

export async function fetchDBTables(conn, group) {
  let client = await clientDataPromise;
  let resp = await client.DataManager.FetchTables({
    body: { conn_name: conn, group }
  });
  return resp.obj.tables || [];
}

export async function LoadDataFromDB(db_conn_name, group, table, target_db) {
  let client = await clientDataPromise;
  let resp = await client.DataManager.LoadDataFromDB({
    body: { db_conn_name, group, table, target_db }
  });
  return (resp.obj.success ? "成功：" : "失败：") + resp.obj.message;
}

export async function uploadData(file) {}
