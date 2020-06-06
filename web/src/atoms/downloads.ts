import { atom, selector } from "recoil";
import { Download, DownloadStatus } from "../services/animuxdData";

export const downloads = atom<Download[]>({ key: "downloads", default: [] });

export const sumCurrentSpeed = selector<number>({
  key: "sumCurrentSpeed",
  get: ({ get }) => {
    const currentDownloads = get(downloads);

    return currentDownloads.reduce(
      (acc, download) => acc + download.CurrentSpeed,
      0
    );
  },
});

export const downloading = selector<Download[]>({
  key: "inProgressDownloads",
  get: ({ get }) => {
    const currentDownloads = get(downloads);

    return currentDownloads.filter((d) => d.Status === DownloadStatus.Downloading);
  },
});


export const waiting = selector<Download[]>({
  key: "waitingDownloads",
  get: ({ get }) => {
    const currentDownloads = get(downloads);

    return currentDownloads.filter((d) => d.Status === DownloadStatus.Waiting);
  },
});
