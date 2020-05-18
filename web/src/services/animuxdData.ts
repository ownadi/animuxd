export const ANIMUXD_API_URL = "http://localhost:1337";

export enum DownloadStatus {
  Waiting = 0,
  Downloading = 1,
  Done = 2,
  Failed = 3,
}

export const DownloadStatusString = {
  [DownloadStatus.Waiting]: "Waiting",
  [DownloadStatus.Downloading]: "Downloading",
  [DownloadStatus.Done]: "Done",
  [DownloadStatus.Failed]: "Failed",
};

export type Download = {
  FileName: string;
  Status: DownloadStatus;
  CurrentSpeed: number;
  AvgSpeed: number;
  Downloaded: number;
  Size: number;
};
