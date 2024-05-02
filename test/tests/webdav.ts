import { createClient } from 'webdav';
import { config } from 'dotenv';
import { createReadStream } from 'fs';
import { basename } from 'path';

config({ path: '../../.env' });

const client = createClient('http://localhost:8082', {
  username: process.env.WEBDAV_USERNAME,
  password: process.env.WEBDAV_PASSWORD,
});

/**
 * Uploads the given file to the root directory of the test WebDAV service.
 */
export async function uploadFile(path: string) {
  const filename = basename(path);
  console.log('Uploading file to WebDAV:', filename);
  createReadStream(path).pipe(client.createWriteStream(filename));
}
