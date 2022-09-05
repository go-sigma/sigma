import axios from "axios";
import { Link } from "react-router-dom";
import React, { Fragment, useEffect, useState } from "react";

import Menu from "../../components/Menu";
import Header from "../../components/Header";
import Footer from "../../components/Footer";

import TableItem from "./TableItem";

import IProject from "../../interfaces/IProject"
import { Helmet } from "react-helmet";

export default function Home() {
  let [projectList, setProjectList] = useState<IProject[]>([]);

  useEffect(() => {
    axios.get('http://localhost:3001/projects')
      .then((response) => {
        if (response.status === 200) {
          setProjectList(response.data as IProject[]);
        }
      });
  }, []);

  return (
    <Fragment>
      <Helmet>
        <title>XImager - Home</title>
      </Helmet>
      <div className="h-screen flex overflow-hidden bg-white">
        <Menu />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="flex-1 relative z-0 overflow-y-auto focus:outline-none" tabIndex={0}>
            <Header title="Home" props={
              <>
                <Link to="editor" className="order-0 inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3">Create</Link>
              </>
            } />
            <div className="hidden mt-1 sm:block">
              <div className="align-middle inline-block min-w-full border-b border-gray-200">
                <table className="min-w-full">
                  <thead>
                    <tr className="border-t border-gray-200">
                      <th className="px-6 py-3 border-b border-gray-200 bg-gray-50 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        <span className="lg:pl-2">XImager</span>
                      </th>
                      <th className="hidden md:table-cell px-6 py-3 border-b border-gray-200 bg-gray-50 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Last updated
                      </th>
                      <th className="pr-6 py-3 border-b border-gray-200 bg-gray-50 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"></th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-100">
                    {
                      projectList.map(m => {
                        return (
                          <TableItem key={m.name} name={m.name} description={m.description} updated={m.updated} />
                        );
                      })
                    }
                  </tbody>
                </table>
              </div>
            </div>
          </main>
          <Footer />
        </div>
      </div>
    </Fragment >
  )
}
