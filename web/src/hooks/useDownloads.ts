import { useEffect, useState } from "react";
import { Download } from "../services/animuxdData";
import { getDownloads } from "../services/animuxd";

const useDownloads = (): Download[] => {
  const [downloads, setDownloads] = useState<Download[]>([]);

  useEffect(() => {
    const fetchDownloads = () => getDownloads().then(setDownloads);

    fetchDownloads();
    const interval = setInterval(fetchDownloads, 1000);

    return (): void => clearInterval(interval);
  }, []);

  return downloads;
};

export default useDownloads;
