import React from 'react';
import styles from "./Tech.module.scss";
import image from '@site/static/img/architecture.drawio.png'

const Tech = () => {
    return (
        <div className={styles.main}>
            <div className={styles.mainContent}>

                <div className={styles.description}>You don't need sudo anymore.</div>
                <div className={styles.title}>Rootless & linux-native</div>

                <div className={styles.subtitle}>Manage VMs per user based on Qemu and Libvirt.</div>
            </div>

            <img src={image} height={"450vh"} alt="showcase"/>
        </div>
    );
};

export default Tech;
