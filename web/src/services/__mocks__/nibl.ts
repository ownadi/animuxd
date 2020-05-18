import { NiblPackage, NiblBot } from "../niblData";

export const search = (query: string): Promise<NiblPackage[]> => {
  return Promise.resolve<NiblPackage[]>([
    {
      botId: 1,
      episodeNumber: 1,
      lastModified: "2020-02-20 21:37:00",
      name: `${query} 01`,
      number: 1337,
      size: "100M",
      sizekbits: 102400,
    },
    {
      botId: 1,
      episodeNumber: 2,
      lastModified: "2020-02-20 21:37:00",
      name: `${query} 02`,
      number: 1338,
      size: "150M",
      sizekbits: 153600,
    },
  ]);
};

export const getBots = (): Promise<NiblBot[]> => {
  return Promise.resolve<NiblBot[]>([
    {
      id: 1,
      batchEnable: 1,
      lastProcessed: "2020-02-20 21:37:00",
      name: "fo0b0t",
      owner: "baz",
      packSize: 1337,
    },
  ]);
};
