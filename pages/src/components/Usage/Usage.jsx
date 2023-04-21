import React from 'react';
import styles from "./Usage.module.scss";
import image from '@site/static/img/shell.png'
// import Prompt from "./Prompt";

const Usage = () => {
    const data = [
        {
            type: 'input',
            value: 'freyja machine create -c examples/myconf.yaml',
            delay: 1000,
        },
        {
            color: '#4bfcd2',
            value: 'Create host vm1 ...',
            delay: 1000,
        },
        {
            color: '#4bfcd2',
            value: 'Create host vm2 ...',
            delay: 1000,
        },
        {
            color: '#4bfcd2',
            value: 'Domain creation completed.',
        },
        { value: '' },
        {
            type: 'input',
            value: 'freyja machine info',
        },
        {
            color: '#4bfcd2',
            value: "vm1:<br/><br/>" +
                "  memory: 4.0 GB\n" +
                "  networks:\n" +
                "  - interface: unknown\n" +
                "    ip: 192.168.122.91/24\n" +
                "    mac: 52:54:00:4f:1f:02\n" +
                "    name: default\n" +
                "    type: network\n" +
                "  state: running\n" +
                "  vcpus: 2\n" +
                "vm2:\n" +
                "  memory: 4.0 GB\n" +
                "  networks:\n" +
                "  - interface: unknown\n" +
                "    ip: 192.168.122.191/24\n" +
                "    mac: 52:54:00:b5:7f:bd\n" +
                "    name: default\n" +
                "    type: network\n" +
                "  state: running\n" +
                "  vcpus: 2",
        },
    ]

    return (
        <div className={styles.main}>
            {/*<Prompt data={data} height={'50vh'} width={'40vw'} />*/}
            <img src={image} height={"350vh"} alt="showcase"/>

            <div className={styles.mainContent}>

                <div className={styles.description}>Stick to your terminal.</div>
                <div className={styles.title}>Shell cli</div>

                <div className={styles.subtitle}>Suitable for developers and integrators.</div>
            </div>

        </div>
    );
};

export default Usage;
