import React, { useMemo } from "react";
import { useTable, Column, useSortBy } from "react-table";
import useDownloads from "../hooks/useDownloads";
import * as Table from "../styles/table";
import { FaArrowDown, FaArrowUp } from "react-icons/fa";
import filesize from "filesize";
import { Download, DownloadStatusString } from "../services/animuxdData";
import styled from "../styles/styled";

const Container = styled.div`
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
`;

const Downloads = () => {
  const downloads = useDownloads();

  const columns = useMemo<Column<Download>[]>(() => {
    return [
      {
        Header: "File",
        accessor: "FileName",
      },
      {
        Header: "Status",
        accessor: "Status",
        Cell: (cell) => DownloadStatusString[cell.value],
      },
      {
        Header: "Progress",
        accessor: "Downloaded",
        Cell: (cell) =>
          (
            (cell.row.original.Downloaded / cell.row.original.Size) *
            100
          ).toFixed(2) + "%",
      },
      {
        Header: "Current speed",
        accessor: "CurrentSpeed",
        Cell: (cell) => filesize(cell.value) + "/s",
      },
      {
        Header: "Avg. speed",
        accessor: "AvgSpeed",
        Cell: (cell) => filesize(cell.value) + "/s",
      },
      {
        Header: "Size",
        accessor: "Size",
        Cell: (cell) => filesize(cell.value),
      },
    ];
  }, []);

  const {
    getTableProps,
    getTableBodyProps,
    headerGroups,
    rows,
    prepareRow,
  } = useTable<Download>(
    {
      columns,
      data: downloads,
      autoResetSortBy: false,
    },
    useSortBy
  );

  return (
    <Container>
      <Table.Container>
        <Table.Table {...getTableProps()}>
          <thead>
            {headerGroups.map((headerGroup) => (
              <tr {...headerGroup.getHeaderGroupProps()}>
                {headerGroup.headers.map((column) => (
                  <Table.Header
                    {...column.getHeaderProps(column.getSortByToggleProps())}
                  >
                    {column.render("Header")}
                    &nbsp;
                    {column.isSorted ? (
                      column.isSortedDesc ? (
                        <FaArrowDown fontSize="1rem" />
                      ) : (
                        <FaArrowUp fontSize="1rem" />
                      )
                    ) : null}
                  </Table.Header>
                ))}
              </tr>
            ))}
          </thead>
          <tbody {...getTableBodyProps()}>
            {rows.map((row) => {
              prepareRow(row);
              return (
                <Table.Row {...row.getRowProps()} data-testid="downloadsRow">
                  {row.cells.map((cell) => {
                    return (
                      <Table.Data {...cell.getCellProps()}>
                        {cell.render("Cell")}
                      </Table.Data>
                    );
                  })}
                </Table.Row>
              );
            })}
          </tbody>
        </Table.Table>
      </Table.Container>
    </Container>
  );
};

export default Downloads;
