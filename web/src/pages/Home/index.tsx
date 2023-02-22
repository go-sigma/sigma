import axios from "axios";
import { Fragment, useEffect, useState } from "react";
import { Dialog, Transition } from '@headlessui/react';
import { Helmet } from "react-helmet";

import Menu from "../../components/Menu";
import Header from "../../components/Header";
import Footer from "../../components/Footer";

import TableItem from "./TableItem";
import Notification from "./Notification";

import { INamespace, INamespaceList, INotification, IHTTPError } from "../../interfaces/interfaces";

export default function Home({ localServer }: { localServer: string }) {
  let [projectList, setProjectList] = useState<INamespaceList>({} as INamespaceList);
  let [namespaceText, setNamespaceText] = useState("");
  let [descriptionText, setDescriptionText] = useState("");
  let [refresh, setRefresh] = useState({});

  let [notification, setNotification] = useState(false);
  let [notificationText, setNotificationText] = useState({} as INotification);

  const [open, setOpen] = useState(false);

  useEffect(() => {
    axios.get(localServer + '/namespace/')
      .then((response) => {
        if (response.status === 200) {
          setProjectList(response.data as INamespaceList);
        }
      });
  }, [refresh]);

  let createNamespace = (namespace: string, description: string) => {
    setOpen(false);
    axios.post(localServer + '/namespace/', {
      name: namespace,
      description: description,
    } as INamespace, {}).then(response => {
      setNamespaceText("");
      console.log(response);
      if (response.status === 201) {
        setRefresh({});
      } else {
        let errorcode = response.data as IHTTPError;
        setNotificationText({ title: errorcode.title, message: errorcode.message });
        setNotification(true);
      }
    }).catch(error => {
      let errorcode = error.response.data as IHTTPError;
      console.log(errorcode);
      setNotificationText({ title: errorcode.title, message: errorcode.message });
      setNotification(true);
    })
  }

  return (
    <Fragment>
      <Helmet>
        <title>XImager - Home</title>
      </Helmet>
      <div className="h-screen flex overflow-hidden bg-white">
        <Menu />
        <div className="flex flex-col w-0 flex-1 overflow-hidden">
          <main className="flex-1 relative z-0 focus:outline-none" tabIndex={0}>
            <Header title="Home" props={
              <button className="order-0 inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 sm:order-1 sm:ml-3"
                onClick={() => {
                  setOpen(true)
                }}
              >Create</button>
            } />
            <div className="hidden mt-1 sm:block">
              <div className="align-middle inline-block min-w-full border-b border-gray-200">
                <table className="min-w-full">
                  <thead>
                    <tr className="border-t border-gray-200">
                      <th className="px-6 py-3 border-b border-gray-200 bg-gray-50 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        <span className="lg:pl-2">Namespace</span>
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
                      projectList.items?.map(m => {
                        return (
                          <TableItem key={m.id} name={m.name} description={m.description} created_at={m.created_at} updated_at={m.updated_at} />
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
      <Transition.Root show={open} as={Fragment}>
        <Dialog as="div" className="relative z-10" onClose={setOpen}>
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
          </Transition.Child>

          <div className="fixed inset-0 z-10 overflow-y-auto">
            <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
              <Transition.Child
                as={Fragment}
                enter="ease-out duration-300"
                enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                enterTo="opacity-100 translate-y-0 sm:scale-100"
                leave="ease-in duration-200"
                leaveFrom="opacity-100 translate-y-0 sm:scale-100"
                leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              >
                <Dialog.Panel className="relative transform overflow-hidden rounded-lg bg-white px-4 pt-5 pb-4 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6">
                  <div className="col-span-6 sm:col-span-3">
                    <label htmlFor="first-name" className="block text-sm font-medium text-gray-700">
                      Namespace
                    </label>
                    <input
                      type="text"
                      name="namespace"
                      placeholder="2-20 characters"
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                      value={namespaceText}
                      onChange={(e) => {
                        setNamespaceText(e.target.value);
                      }}
                    />
                  </div>
                  <div className="col-span-6 sm:col-span-3 mt-5">
                    <label htmlFor="first-name" className="block text-sm font-medium text-gray-700">
                      Description
                    </label>
                    <input
                      type="text"
                      name="description"
                      placeholder="30 characters"
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                      value={descriptionText}
                      onChange={e => setDescriptionText(e.target.value)}
                    />
                  </div>
                  <div className="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                    <button
                      type="button"
                      className="inline-flex w-full justify-center rounded-md border border-transparent bg-indigo-500 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:bg-indigo-500 focus:ring-offset-2 sm:ml-3 sm:w-auto sm:text-sm"
                      onClick={() => createNamespace(namespaceText, descriptionText)}
                    >
                      Create
                    </button>
                    <button
                      type="button"
                      className="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:mt-0 sm:w-auto sm:text-sm"
                      onClick={() => setOpen(false)}
                    >
                      Cancel
                    </button>
                  </div>
                </Dialog.Panel>
              </Transition.Child>
            </div>
          </div>
        </Dialog>
      </Transition.Root>
      {
        notification ? <Notification initShow={notification} notificationText={notificationText} /> : null
      }
    </Fragment >
  )
}
