export const NIBL_API_URL = "https://api.nibl.co.uk/nibl";

export type NiblPackage = {
  botId: number;
  number: number;
  name: string;
  size: string;
  sizekbits: number;
  episodeNumber: number;
  lastModified: string;
};

export type NiblBot = {
  id: number;
  name: string;
  owner: string;
  lastProcessed: string;
  batchEnable: number;
  packSize: number;
};
