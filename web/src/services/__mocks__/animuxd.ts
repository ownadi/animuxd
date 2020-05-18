import { NiblPackage } from "../niblData";
import { Download, DownloadStatus } from "../animuxdData";

export const requestFile = async (niblPackage: NiblPackage): Promise<void> => {
  return Promise.resolve<void>(undefined);
};

export const getDownloads = jest.fn(
  (): Promise<Download[]> => {
    return Promise.resolve<Download[]>([
      {
        FileName: "foo.mkv",
        Downloaded: (1024 * 1024 * 1024) / 2,
        Size: 1024 * 1024 * 1024,
        Status: DownloadStatus.Downloading,
        AvgSpeed: 1024 * 1024 * 3,
        CurrentSpeed: 1024 * 1024 * 10,
      },
      {
        FileName: "bar.mkv",
        Downloaded: 0,
        Size: 2048 * 1024 * 1024,
        Status: DownloadStatus.Waiting,
        AvgSpeed: 0,
        CurrentSpeed: 0,
      },
    ]);
  }
);
