import { Link, Outlet } from 'umi';
import {NextUIProvider} from "@nextui-org/react";

export default function Layout() {
  return (
    <NextUIProvider>
      <Outlet />
    </NextUIProvider>
  );
}
