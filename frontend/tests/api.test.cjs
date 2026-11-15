const { test, afterEach } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const Module = require('node:module');
const path = require('node:path');
const ts = require('typescript');
const axios = require('axios');

const filename = path.resolve(__dirname, '../src/app/api.ts');
const compiled = ts.transpileModule(fs.readFileSync(filename, 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, esModuleInterop: true } });
const loaded = new Module(filename, module);
loaded.filename = filename;
loaded.paths = Module._nodeModulePaths(path.dirname(filename));
loaded._compile(compiled.outputText, filename);
const { api, setAccessToken } = loaded.exports;
const adapter = api.defaults.adapter;
afterEach(() => { setAccessToken(null); api.defaults.adapter = adapter; });

test('attaches the current token and removes it on logout', async () => {
  const seen = [];
  api.defaults.adapter = async config => { seen.push(config.headers.get('Authorization')); return { data: {}, status: 200, statusText: 'OK', headers: {}, config }; };
  setAccessToken('first');
  await api.get('monitors');
  setAccessToken('second');
  await api.get('monitors');
  setAccessToken(null);
  await api.get('monitors');
  assert.deepEqual(seen, ['Bearer first', 'Bearer second', undefined]);
});

test('preserves API error messages and cancellation', async () => {
  api.defaults.adapter = async config => { throw new axios.AxiosError('failed', 'ERR_BAD_REQUEST', config, null, { status: 400, data: { error: 'invalid input' }, headers: {}, config }); };
  await assert.rejects(api.get('monitors'), { message: 'invalid input' });
  const controller = new AbortController();
  controller.abort();
  await assert.rejects(api.get('monitors', { signal: controller.signal }), error => axios.isCancel(error));
});

test('rejects malformed successful responses and accepts empty 204', async () => {
  api.defaults.adapter = async config => ({ data: 'invalid', status: 200, statusText: 'OK', headers: {}, config });
  await assert.rejects(api.get('monitors'), { message: 'Сервер вернул некорректный ответ' });
  api.defaults.adapter = async config => ({ data: '', status: 204, statusText: 'No Content', headers: {}, config });
  assert.equal((await api.post('auth/logout')).status, 204);
});
