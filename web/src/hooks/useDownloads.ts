import { useEffect } from "react";
import { Download } from "../services/animuxdData";
import { getDownloads } from "../services/animuxd";
import isEqual from "fast-deep-equal";
import { useSetRecoilState } from "recoil";
import { downloads as downloadsAtom } from "../atoms/downloads";

/**
 * Populates recoils's atom with fresh downloads every second.
 */
const useDownloads = (): void => {
  const setDownloads = useSetRecoilState(downloadsAtom);

  useEffect(() => {
    let previousDownloads: Download[] = [];

    const fetchDownloads = () =>
      getDownloads().then((newDownloads) => {
        const sorted = newDownloads.sort((a, b) => a.FileName.localeCompare(b.FileName)); // TODO: sort by start time
        if (isEqual(previousDownloads, newDownloads)) return;

        previousDownloads = sorted;
        setDownloads(sorted);
      });

    fetchDownloads();
    const interval = setInterval(fetchDownloads, 1000);

    return (): void => clearInterval(interval);
  }, [setDownloads]);
};

export default useDownloads;
