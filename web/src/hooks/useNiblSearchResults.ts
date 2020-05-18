import { useEffect, useState } from "react";
import { NiblPackage } from "../services/niblData";
import { search } from "../services/nibl";

let latest: NiblPackage[] = [];
let latestQuery = "";

type Result = {
  loading: boolean;
  query: string;
  searchResults: NiblPackage[];
};

const useNiblSearchResults = (query: string | null): Result => {
  const [result, setResult] = useState<Result>({
    loading: false,
    query: latestQuery,
    searchResults: latest,
  });

  useEffect(() => {
    if (!query) {
      return;
    }

    setResult((r) => ({ ...r, loading: true, query: "" }));

    search(query).then((packages) => {
      setResult({ loading: false, query, searchResults: packages });
      latest = packages;
      latestQuery = query;
    });
  }, [query]);

  return result;
};

export default useNiblSearchResults;
