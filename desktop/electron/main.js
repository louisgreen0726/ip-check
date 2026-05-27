const { app, BrowserWindow, dialog, shell } = require("electron");
const { spawn } = require("node:child_process");
const fs = require("node:fs");
const http = require("node:http");
const net = require("node:net");
const path = require("node:path");

let mainWindow = null;
let serverProcess = null;
let serverURL = "";

const gotLock = app.requestSingleInstanceLock();
if (!gotLock) {
  app.quit();
}

app.on("second-instance", () => {
  if (mainWindow) {
    if (mainWindow.isMinimized()) mainWindow.restore();
    mainWindow.focus();
  }
});

app.whenReady().then(async () => {
  try {
    const port = await findFreePort();
    serverURL = `http://127.0.0.1:${port}`;
    await startBackend(port);
    await waitForHealth(`${serverURL}/api/health`, 8000);
    createWindow();
  } catch (error) {
    dialog.showErrorBox("IP Check", error && error.message ? error.message : String(error));
    app.quit();
  }
});

app.on("before-quit", () => {
  stopBackend();
});

app.on("window-all-closed", () => {
  app.quit();
});

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1360,
    height: 920,
    minWidth: 980,
    minHeight: 720,
    title: "IP Check",
    backgroundColor: "#fbfbfa",
    webPreferences: {
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true
    }
  });

  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url);
    return { action: "deny" };
  });

  mainWindow.loadURL(serverURL);
}

async function startBackend(port) {
  const binary = prepareBackendBinary(resolveBackendBinary());
  if (!fs.existsSync(binary)) {
    throw new Error(`找不到内置后端二进制: ${binary}`);
  }
  serverProcess = spawn(binary, ["serve", "--addr", `127.0.0.1:${port}`], {
    stdio: ["ignore", "pipe", "pipe"],
    env: {
      ...process.env,
      IP_CHECK_DESKTOP: "1"
    }
  });

  serverProcess.stdout.on("data", (chunk) => {
    console.log(String(chunk).trim());
  });
  serverProcess.stderr.on("data", (chunk) => {
    console.error(String(chunk).trim());
  });
  serverProcess.on("exit", (code, signal) => {
    if (mainWindow && !mainWindow.isDestroyed()) {
      console.error(`ipcheck backend exited: code=${code} signal=${signal}`);
    }
  });
}

function stopBackend() {
  if (serverProcess && !serverProcess.killed) {
    serverProcess.kill("SIGTERM");
  }
  serverProcess = null;
}

function resolveBackendBinary() {
  if (process.env.IPCHECK_BINARY) {
    return process.env.IPCHECK_BINARY;
  }
  if (app.isPackaged) {
    return path.join(process.resourcesPath, "ipcheck", "ipcheck");
  }
  return path.resolve(__dirname, "../../bin/ipcheck");
}

function prepareBackendBinary(binary) {
  if (process.platform === "win32") {
    return binary;
  }
  try {
    fs.accessSync(binary, fs.constants.X_OK);
    return binary;
  } catch (error) {
    if (!app.isPackaged) {
      fs.chmodSync(binary, 0o755);
      return binary;
    }
  }

  const targetDir = path.join(app.getPath("userData"), "backend");
  const target = path.join(targetDir, path.basename(binary));
  fs.mkdirSync(targetDir, { recursive: true });
  fs.copyFileSync(binary, target);
  fs.chmodSync(target, 0o755);
  return target;
}

function findFreePort() {
  return new Promise((resolve, reject) => {
    const server = net.createServer();
    server.unref();
    server.on("error", reject);
    server.listen(0, "127.0.0.1", () => {
      const address = server.address();
      server.close(() => resolve(address.port));
    });
  });
}

function waitForHealth(url, timeoutMs) {
  const started = Date.now();
  return new Promise((resolve, reject) => {
    const attempt = () => {
      const req = http.get(url, (res) => {
        res.resume();
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve();
          return;
        }
        retry();
      });
      req.on("error", retry);
      req.setTimeout(800, () => {
        req.destroy();
        retry();
      });
    };

    const retry = () => {
      if (Date.now() - started > timeoutMs) {
        reject(new Error("后端启动超时"));
        return;
      }
      setTimeout(attempt, 150);
    };

    attempt();
  });
}
