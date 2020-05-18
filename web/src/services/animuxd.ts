import { ANIMUXD_API_URL, Download } from "./animuxdData";
import { NiblPackage, NiblBot } from "./niblData";
import { getBots } from "./nibl";

let botsCache: NiblBot[] = [];

export const requestFile = async (niblPackage: NiblPackage): Promise<void> => {
  let botsPromise: Promise<NiblBot[]>;

  if (botsCache.length === 0) {
    botsPromise = getBots();
  } else {
    botsPromise = Promise.resolve(botsCache);
    botsPromise.then((bots) => (botsCache = bots));
  }

  const bots = await botsPromise;
  const bot = bots.find((b) => b.id === niblPackage.botId);
  if (!bot) return;

  fetch(`${ANIMUXD_API_URL}/downloads`, {
    method: "POST",
    body: JSON.stringify({
      botNick: bot.name,
      packageNumber: niblPackage.number,
      fileName: niblPackage.name,
    }),
  });
};

export const getDownloads = (): Promise<Download[]> => {
  return fetch(`${ANIMUXD_API_URL}/downloads`, {
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
    },
  }).then((response) => response.json() as Promise<Download[]>);
};
