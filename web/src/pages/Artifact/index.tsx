import axios from "axios";
import { useParams } from 'react-router-dom';
import { Fragment, useEffect, useState } from "react";
import { Helmet, HelmetProvider } from 'react-helmet-async';
import { useSearchParams } from 'react-router-dom';

import Menu from "../../components/Menu";
import Header from "../../components/Header";
import Pagination from "../../components/Pagination";
import Settings from "../../Settings";

import TableItem from "./TableItem";
import "./index.css";

import { IArtifact, IArtifactList, IHTTPError } from "../../interfaces/interfaces";

export default function Artifact({ localServer }: { localServer: string }) {
  let [artifactList, setArtifactList] = useState<IArtifactList>({} as IArtifactList);
  let [refresh, setRefresh] = useState({});
  let [pageNum, setPageNum] = useState(1);
  let [total, setTotal] = useState(0);

  const { namespace } = useParams<{ namespace: string }>();
  const [searchParams] = useSearchParams();
  const repository = searchParams.get('repository');

  useEffect(() => {
    let url = localServer + `/namespace/${namespace}/artifact/?repository=${repository}&page_size=${Settings.PageSize}&page_num=${pageNum}`;
    axios.get(url)
      .then((response) => {
        if (response.status === 200) {
          let artifactList = response.data as IArtifactList;
          setArtifactList(artifactList);
          setTotal(artifactList.total);
        }
      });
  }, [refresh, pageNum]);

  return (
    <Fragment>
      <HelmetProvider>
        <Helmet>
          <title>XImager - Artifact</title>
        </Helmet>
      </HelmetProvider>
      <div className="min-h-screen flex overflow-hidden bg-white">
        <Menu item="Artifact" />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="flex-1 relative z-0 focus:outline-none" tabIndex={0}>
            <Header title="Artifact" />
            <div className="hidden sm:block">
              <div className="align-middle inline-block min-w-full border-b border-gray-200">
                <table className="min-w-full">
                  <thead>
                    <tr className="border-gray-200">
                      <th className="px-6 py-3 border-b border-gray-200 bg-gray-50 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        <span className="lg:pl-2">Digest</span>
                      </th>
                      <th className="hidden md:table-cell px-6 py-3 border-b border-gray-200 bg-gray-50 text-right text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                        Tags
                      </th>
                      <th className="hidden md:table-cell px-6 py-3 border-b border-gray-200 bg-gray-50 text-right text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                        Tag Count
                      </th>
                      <th className="hidden md:table-cell px-6 py-3 border-b border-gray-200 bg-gray-50 text-right text-xs font-medium text-gray-500 uppercase tracking-wider whitespace-nowrap">
                        Size
                      </th>
                      <th className="hidden md:table-cell px-6 py-3 border-b border-gray-200 bg-gray-50 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Create
                      </th>
                      <th className="hidden md:table-cell px-6 py-3 border-b border-gray-200 bg-gray-50 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Update
                      </th>
                      <th className="pr-6 py-3 border-b border-gray-200 bg-gray-50 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Action</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-100">
                    {
                      artifactList.items?.map(m => {
                        return (
                          <TableItem key={m.id} id={m.id} namespace={namespace} repository={repository} digest={m.digest} size={m.size} tags={m.tags} tag_count={m.tag_count} created_at={m.created_at} updated_at={m.updated_at} />
                        );
                      })
                    }
                  </tbody>
                </table>
              </div>
            </div>
          </main>
          <Pagination page_size={Settings.PageSize} page_num={pageNum} setPageNum={setPageNum} total={total} />
        </div>
      </div>
    </Fragment >
  )
}
