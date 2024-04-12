import Dashboard from "./views/Dashboard.js";
import PhishingStudio from "./views/PhishingStudio.js";
import PayloadStudio from "./views/PayloadStudio.js";
import Documentation from "./views/Documentation.js";

var routes = [
  {
    path: "/dashboard",
    name: "C2 Studio",
    icon: "tim-icons icon-laptop",
    component: <Dashboard />,
    layout: "/admin",
  },
  {
    path: "/phishing",
    name: "Phishing Studio",
    icon: "tim-icons icon-email-85",
    component: <PhishingStudio />,
    layout: "/admin",
  },
  {
    path: "/payload",
    name: "Payload Studio",
    icon: "tim-icons icon-settings",
    component: <PayloadStudio />,
    layout: "/admin",
  }
];
export default routes;
